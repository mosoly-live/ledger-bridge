package main

import (
	"context"
	"database/sql"
	"expvar"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"gitlab.com/p-invent/mosoly-ledger-bridge/mosolyapi"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jmoiron/sqlx"
	"github.com/monetha/go-distributed"
	"github.com/monetha/go-ethereum/blocksource"
	"gitlab.com/p-invent/mosoly-ledger-bridge/config"
	"gitlab.com/p-invent/mosoly-ledger-bridge/log"
	_ "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/db/pqtimeouts"
	"gitlab.com/p-invent/mosoly-ledger-bridge/processing/txnprocessing"
	"gitlab.com/p-invent/mosoly-ledger-bridge/processing/txnvalidating"
	"gitlab.com/p-invent/mosoly-ledger-bridge/repository"
	"gitlab.com/p-invent/mosoly-ledger-bridge/restapi"
)

const sqlDriverName = "pq-timeouts"

var (
	// Version (set by compiler) is the version of program in YYYY.MM.DD.PILELINE_ID format
	Version = "undefined"
	// BuildTime (set by compiler) is the program build time in '+%Y-%m-%dT%H:%M:%SZ' format
	BuildTime = "undefined"
	// GitHash (set by compiler) is the git commit hash of source tree
	GitHash = "undefined"
	// Branch (set by compiler) is the name of branch of source tree
	Branch = "undefined"
	// DefaultHTTPClient is the default HTTP client we use for external calls.
	DefaultHTTPClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			MaxIdleConnsPerHost:   100, // default is 2, we've increased it
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

func init() {
	expvar.NewString("Version").Set(Version)
	expvar.NewString("GitHash").Set(GitHash)
	expvar.NewString("BuildTime").Set(BuildTime)
}

func main() {
	if Version == "vlatest" || Version == "undefined" {
		log.ReplaceGlobals(log.NewDevelopment())
	} else {
		log.ReplaceGlobals(log.NewProduction())
	}

	defer log.Sync()
	defer log.Println("service fully stopped")

	log.Println("Version:", Version)
	log.Println("BuildTime:", BuildTime)
	log.Println("GitHash:", GitHash)
	log.Println("Branch:", Branch)

	config.Parse()
	log.Println("running service...")
	if err := run(); err != nil {
		log.Fatalf("error: %v\n", err)
	}
}

func run() error {
	ctx := createTerminationContext()

	log.Println("sql.Open...")
	sdb, err := sql.Open(sqlDriverName, config.SQLConnectionString)
	sqlxdb := sqlx.NewDb(sdb, sqlDriverName)
	if err != nil {
		return fmt.Errorf("sql.Open: %v", err)
	}

	log.Println("repository.New...")
	repo, err := repository.New(sdb, sqlDriverName)
	if err != nil {
		return fmt.Errorf("new repository: %v", err)
	}
	defer logClose(repo, "repository")

	log.Println("creating consul broker...")

	// creating broker to run distributed tasks
	br, err := distributed.NewBroker(config.ConsulConfig)
	if err != nil {
		return fmt.Errorf("creating task broker: %v", err)
	}

	// Create long-running task for ethereum processing
	ethclient, err := ethclient.Dial(config.EthereumJSONRPCURL)
	if err != nil {
		return fmt.Errorf("new ethereum client: %v", err)
	}

	log.Println("api client New...")
	apiClient, err := mosolyapi.NewClient(DefaultHTTPClient, config.AppMosolyBackendURL)
	if err != nil {
		return fmt.Errorf("creating client for Mosoly backend: %v", err)
	}

	log.Println("txnprocessing New...")
	txn, err := txnprocessing.New(sqlxdb, ethclient, apiClient, DefaultHTTPClient)
	if err != nil {
		return fmt.Errorf("creating txnprocessing processing instance: %v", err)
	}

	log.Println("creating transaction processing task...")
	txnProcessingTask, err := br.NewTask(path.Join(config.ConsulKeyPrefix, "processing/transaction/task"), txn.Run)
	if err != nil {
		return fmt.Errorf("creating transaction processing long-running task: %v", err)
	}
	defer logClose(txnProcessingTask, "transaction processing task")

	// Create long-running tasks for transaction validation
	log.Println("txnvalidating New...")
	txnValidating, err := txnvalidating.New(repo, blockSourceCreator{config.EthereumJSONRPCURL})
	if err != nil {
		return fmt.Errorf("creating txnvalidating processing instance: %v", err)
	}

	log.Println("creating transaction validating task...")
	txnValidatingTask, err := br.NewTask(path.Join(config.ConsulKeyPrefix, "validating/transaction/task"), txnValidating.Run)
	if err != nil {
		return fmt.Errorf("creating transaction validating long-running task: %v", err)
	}
	defer logClose(txnValidatingTask, "transaction validating task")

	service := restapi.NewService(&restapi.ServiceConfig{
		AllowedOrigins: []string{"*"},
		Port:           config.HTTPPort,
	})

	log.Println("serve HTTP...")
	return service.Serve(ctx)
}

func createTerminationContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("got interrupt signal")
	}()

	return ctx
}

func logClose(c io.Closer, name string) {
	logCloseFun(c.Close, name)
}

func logCloseFun(close func() error, name string) {
	log.Printf("closing %s...", name)
	if err := close(); err != nil {
		log.Printf("close %s: %v", name, err)
	}
}

type blockSourceCreator struct {
	rpcurl string
}

// CreateBlockSource implements processing.BlockSourceCreator
func (c blockSourceCreator) CreateBlockSource(startBlock *big.Int, confirmations uint) (txnvalidating.BlockSource, error) {
	return blocksource.New(c.rpcurl, &blocksource.Config{
		StartBlock:    startBlock,
		Confirmations: confirmations,
	})
}
