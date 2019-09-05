package repository

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"gitlab.com/p-invent/mosoly-ledger-bridge/log"
	"gitlab.com/p-invent/mosoly-ledger-bridge/models/repomodels"
)

// Repository is a main repository to store and load the data
type Repository struct {
	db *sqlx.DB
}

// New creates new instance of Repository
func New(db *sql.DB, driverName string) (repository *Repository, err error) {
	sdb := sqlx.NewDb(db, driverName)
	defer func() {
		if err != nil {
			sdb.Close()
			return
		}
	}()

	log.Println("repository: ping: establishing connection to DB...")
	err = sdb.Ping()
	if err != nil {
		return
	}
	repository = &Repository{
		db: sdb,
	}
	return
}

// UpdateTxnsStatus updates transactions status
func (r *Repository) UpdateTxnsStatus(txHashes []string, status int64, audit repomodels.AuditNameGetter) (err error) {
	db := r.db

	statement := `UPDATE transactions
		SET transaction_state_id = ?,
		updated = timezone('utc', NOW()),
		modified_by = ?
		WHERE transaction_state_id = ? AND transaction_hash IN (?);`
	query, args, err := sqlx.In(statement, status, audit.GetAuditName(), TxnInProgress, txHashes)
	if err != nil {
		return err
	}

	rows, err := db.Query(db.Rebind(query), args...)
	rows.Close()

	return
}

// GetLatestProcessedEthereumBlockNumber returns the latest processed ethereum block number.
func (r *Repository) GetLatestProcessedEthereumBlockNumber(blockNumberID int64, defaultStartBlock uint64) (blockNumber *uint64, err error) {
	db := r.db
	err = db.Get(&blockNumber, db.Rebind(`SELECT latest_processed_block_number
		FROM ethereum_blockchain
		WHERE id = ?`), blockNumberID)

	if err == sql.ErrNoRows {
		blockNumber = &defaultStartBlock
		return blockNumber, nil
	}

	return
}

// SetLatestProcessedEthereumBlockNumber saves the latest processed ethereum block number.
func (r *Repository) SetLatestProcessedEthereumBlockNumber(blockNumberID int64, blockNumber uint64) (err error) {
	db := r.db
	_, err = db.Exec(db.Rebind(`INSERT INTO ethereum_blockchain (id, latest_processed_block_number)
	VALUES (?, ?)
	ON CONFLICT (id) DO UPDATE SET latest_processed_block_number = ?`),
		blockNumberID, blockNumber, blockNumber,
	)
	return
}

// DeleteSuccessfulTransactions deletes successful completed transactions
func (r *Repository) DeleteSuccessfulTransactions() (err error) {
	db := r.db
	_, err = db.Exec(db.Rebind(`DELETE FROM transactions t
		WHERE transaction_state_id = ? and NOT EXISTS (SELECT 1
			FROM user_data u
			WHERE u.transaction_id = t.id) and NOT EXISTS(SELECT 1
			FROM project_data p
			WHERE p.transaction_id = t.id) and NOT EXISTS(SELECT 1
			FROM mentorship m
			WHERE m.transaction_id = t.id)`),
		TxnSuccessful)
	return
}

// Close implements io.Closer
func (r *Repository) Close() error {
	return r.db.Close()
}
