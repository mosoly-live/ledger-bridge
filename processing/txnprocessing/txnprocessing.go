package txnprocessing

import (
	"context"
	"log"
	"net/http"
	"time"

	"gitlab.com/p-invent/mosoly-ledger-bridge/mosolyapi"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jmoiron/sqlx"
)

// MosolyClient is a client that interacts with Mosoly api.
type MosolyClient interface {
	GetProjectUpdates(ctx context.Context, since time.Time) ([]mosolyapi.Project, error)
	GetUserUpdates(ctx context.Context, since time.Time) ([]mosolyapi.User, error)
}

// TxnProcessing for transaction processing
type TxnProcessing struct {
	db         *sqlx.DB
	ethClient  *ethclient.Client
	httpClient *http.Client
	apiClient  MosolyClient
}

// New returns new instance of TxnProcessing
func New(db *sqlx.DB, c *ethclient.Client, apiClient MosolyClient, httpClient *http.Client) (*TxnProcessing, error) {

	return &TxnProcessing{db: db, ethClient: c, httpClient: httpClient, apiClient: apiClient}, nil
}

// GetAuditName audit name
func (t *TxnProcessing) GetAuditName() string {
	return "mosoly-txnprocessing"
}

// Run runs processing synchronously
func (t *TxnProcessing) Run(ctx context.Context) (err error) {
	const notifyTimeout = time.Minute
	now := time.Now().UTC()
	txnProcessRunAt := getTxnProcessRunAt(now)
	var tm *time.Timer // reusable timer
	for {
		if tm == nil {
			tm = time.NewTimer(notifyTimeout)
		} else {
			tm.Reset(notifyTimeout)
		}

		select {
		case <-ctx.Done():
			log.Println("txnprocessing: Service stopped !!!")
			return ctx.Err()
		case tick := <-tm.C: // processing timeout
			now := time.Now().UTC()
			if tick.Unix() >= txnProcessRunAt.Unix() {
				txnProcessRunAt = getTxnProcessRunAt(now)
				log.Println("txnprocessing: updating user and project data to public ledger")
				err := t.processTxns(ctx)
				if err != nil {
					log.Println("txnprocessing: ", err)
				}
			}
		}
	}
}
