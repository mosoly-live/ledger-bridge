package txnvalidating

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	ethereum "github.com/monetha/go-ethereum"
	"gitlab.com/p-invent/mosoly-ledger-bridge/models/repomodels"
	"gitlab.com/p-invent/mosoly-ledger-bridge/repository"
)

const (
	// latest processed block number column id
	blockNumberID = 1
	// number of confirmations for delivered blocks in Ethereum block-chain
	confirmations = 2
)

var (
	// default start block if not set in db
	defaultStartBlock = uint64(4029220)
)

// Repository has methods for database operations.
type Repository interface {
	UpdateTxnsStatus(txHashes []string, status int64, audit repomodels.AuditNameGetter) error
	GetLatestProcessedEthereumBlockNumber(blockNumberID int64, defaultStartBlock uint64) (*uint64, error)
	SetLatestProcessedEthereumBlockNumber(blockNumberID int64, blockNumber uint64) error
	DeleteSuccessfulTransactions() error
}

// TxnValidating for transaction validating
type TxnValidating struct {
	r   Repository
	bsc BlockSourceCreator
}

// New returns new instance of TxnProcessing
func New(r Repository, bsc BlockSourceCreator) (*TxnValidating, error) {
	return &TxnValidating{r: r, bsc: bsc}, nil
}

// GetAuditName audit name
func (t *TxnValidating) GetAuditName() string {
	return "mosoly-txnvalidating"
}

// Run runs processing synchronously
func (t *TxnValidating) Run(ctx context.Context) (err error) {
	latestBlock, err := t.r.GetLatestProcessedEthereumBlockNumber(blockNumberID, defaultStartBlock)
	if err != nil || latestBlock == nil {
		err = fmt.Errorf("txnvalidating: getting latest block number: %v", err)
		return
	}

	// add one, to skip already processed block
	startBlock := big.NewInt(int64(*latestBlock + 1))

	// synchronisation primitives
	var wg sync.WaitGroup
	stopped := make(chan struct{})
	defer func() {
		close(stopped) // notify all goroutines before leaving
		wg.Wait()      // wait for all goroutines are stopped
	}()

	// create the source of blocks
	bs, err := t.bsc.CreateBlockSource(startBlock, confirmations)
	if err != nil {
		err = fmt.Errorf("txnvalidating: creating block source: %v", err)
		return
	}

	// run guard that will close block source on exit, to avoid leak
	wg.Add(1)
	go func() {
		defer wg.Done()

		// wait either for cancellation or for stop
		select {
		case <-ctx.Done():
		case <-stopped:
		}

		// close block source
		if bserr := bs.Close(); bserr != nil && err == nil {
			err = fmt.Errorf("txnvalidating: closing block source: %v", bserr)
		}
	}()

	// process all blocks/transactions from ethereum
	for block := range bs.Blocks() {
		err = t.r.DeleteSuccessfulTransactions()
		if err != nil {
			err = fmt.Errorf("txnvalidating: deleting successful completed transactions: %v", err)
			return
		}

		log.Printf("txnvalidating: processing block: %v", block.Number)
		err = t.validateBlockTxns(block)
		if err != nil {
			err = fmt.Errorf("txnvalidating: validate block transactions: %v", err)
			return
		}

		t.r.SetLatestProcessedEthereumBlockNumber(blockNumberID, uint64(block.Number.Int64()))
	}

	return
}

func (t *TxnValidating) validateBlockTxns(block *ethereum.Block) (err error) {
	// proccess successful transactions
	txsHashSuccessful := getTxHashesByStatus(block.Transactions, ethereum.TransactionSuccessful)
	if len(txsHashSuccessful) > 0 {
		err = t.r.UpdateTxnsStatus(txsHashSuccessful, repository.TxnSuccessful, t)
		if err != nil {
			err = fmt.Errorf("txnvalidating: error in validating successful transactions: %v", err)
			return
		}
	}

	// proccess failed transactions
	txsHashFailed := getTxHashesByStatus(block.Transactions, ethereum.TransactionFailed)
	if len(txsHashFailed) > 0 {
		err = t.r.UpdateTxnsStatus(txsHashFailed, repository.TxnFailed, t)
		if err != nil {
			err = fmt.Errorf("txnvalidating: error in failing transactions: %v", err)
			return
		}
	}

	return
}
