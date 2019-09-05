package txnprocessing

import (
	"context"
	"fmt"

	"gitlab.com/p-invent/mosoly-ledger-bridge/repository"
)

func (t *TxnProcessing) processTxns(ctx context.Context) (err error) {
	updatedUsers, err := t.getUpdates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get recently updated users: %v", err)
	}

	projects, err := t.syncProjects(ctx)
	if err != nil {
		return err
	}

	users, err := t.syncUsers(ctx, updatedUsers)
	if err != nil {
		return err
	}

	err = t.syncToBlockchain(ctx, projects, users)
	if err != nil {
		return err
	}

	return
}

func (t *TxnProcessing) updateUserFactTransaction(userID, trxID int) error {
	db := t.db
	rows, err := db.Query(db.Rebind(`UPDATE user_data SET
	transaction_id = ?
	WHERE id = ?;`), trxID, userID)

	if err != nil {
		return err
	}

	rows.Close()

	return nil
}

func (t *TxnProcessing) updateMentorFactTransaction(userID, trxID int) error {
	db := t.db
	rows, err := db.Query(db.Rebind(`UPDATE mentorship SET
	transaction_id = ?
	WHERE user_id = ?;`), trxID, userID)

	if err != nil {
		return err
	}

	rows.Close()

	return nil
}

func (t *TxnProcessing) updateProjectFactTransaction(projectID, trxID int) error {
	db := t.db
	rows, err := db.Query(db.Rebind(`UPDATE project_data SET
	transaction_id = ?
	WHERE id = ?;`), trxID, projectID)

	if err != nil {
		return err
	}

	rows.Close()

	return nil
}

func (t *TxnProcessing) createTxnData(hash string) (id int, err error) {
	db := t.db
	rows, err := db.Query(db.Rebind(`INSERT INTO transactions (
		created,
		updated,
		modified_by,
		transaction_hash,
		transaction_state_id)
		VALUES(timezone('utc',NOW()), timezone('utc',NOW()), ?, ?, ?)
		RETURNING id`),
		t.GetAuditName(), hash, repository.TxnInProgress)
	if err != nil {
		return
	}

	if rows.Next() {
		err = rows.Scan(&id)
		rows.Close()
		if err != nil {
			return
		}
	} else {
		panic("id of transaction expected")
	}

	return
}
