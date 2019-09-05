package txnprocessing

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/p-invent/mosoly-ledger-bridge/mosolyapi"

	"gitlab.com/p-invent/mosoly-ledger-bridge/models/dbmodels"

	"gitlab.com/p-invent/mosoly-ledger-bridge/models/transformations"
)

var (
	syncFallbackDate = time.Unix(0, 0)
)

func (t *TxnProcessing) getUpdates(ctx context.Context) ([]mosolyapi.User, error) {
	maxUpdateDate, err := t.getMaxUpdateDate("user_data")
	if err != nil {
		return nil, err
	}
	fmt.Println(maxUpdateDate)

	// Get users updated after max update time.
	users, err := t.apiClient.GetUserUpdates(ctx, *maxUpdateDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get recently updated users: %v", err)
	}

	return users, nil
}

func (t *TxnProcessing) syncUsers(ctx context.Context, users []mosolyapi.User) ([]*dbmodels.User, error) {
	db := t.db

	tx, err := db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin user insert/update transaction for user_data: %v", err)
	}
	defer tx.Rollback()

	dbUsers := make([]*dbmodels.User, 0)

	// Update or insert each user.
	for _, us := range users {
		user := transformations.TransformUser(&us)

		var notExists bool

		err = tx.QueryRow(tx.Rebind(`
			SELECT id FROM user_data
			WHERE id = ?
			ORDER BY id DESC LIMIT 1`), user.ID,
		).Scan(&user.ID)

		dbUsers = append(dbUsers, user)

		if err != nil {
			if err == sql.ErrNoRows {
				notExists = true
			} else {
				return nil, fmt.Errorf("failed to get user ID: %v", err)
			}
		}

		if notExists {
			err = tx.QueryRow(tx.Rebind(`
				INSERT INTO user_data(
					id, invite_url_hash, account, updated_at, validated
				) VALUES (?, ?, ?, ?, ?)
				RETURNING id`), user.ID, user.InviteURLHash, user.Account, user.UpdatedAt, user.Validated).
				Scan(&user.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to insert user: %v", err)
			}
			continue
		}

		// If exists:
		_, err = tx.Exec(tx.Rebind(`
			UPDATE user_data SET
				invite_url_hash = ?,
				account = ?,
				updated_at = ?,
				validated = ?
			WHERE id = ?`), user.InviteURLHash, user.Account, user.UpdatedAt, user.Validated, user.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %v", err)
		}

		_, err = tx.Exec(tx.Rebind(`DELETE FROM mentorship WHERE user_id = ?`),
			user.ID)

		if err != nil {
			return nil, err
		}

		for _, mentoree := range user.Mentorees {
			_, err = tx.Exec(tx.Rebind(`INSERT INTO mentorship(user_id, mentoree_id) VALUES (?, ?)`),
				user.ID, mentoree.ID)

			if err != nil {
				return nil, fmt.Errorf("failed to update mentorship: %v", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit syncUsers transaction: %v", err)
	}

	return dbUsers, nil
}
