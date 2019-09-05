package txnprocessing

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/p-invent/mosoly-ledger-bridge/models/dbmodels"

	"github.com/lib/pq"
)

func (t *TxnProcessing) getMaxUpdateDate(tableName string) (*time.Time, error) {
	db := t.db

	var nullMaxUpdateDate pq.NullTime

	// Get max update time
	err := db.QueryRow(fmt.Sprintf(`SELECT MAX(updated_at) FROM %s`, tableName)).Scan(&nullMaxUpdateDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get max 'updated_at' for project_data: %v", err)
	}

	// Use fallback time if there are no projects yet.
	var maxUpdateDate time.Time
	if nullMaxUpdateDate.Valid {
		maxUpdateDate = nullMaxUpdateDate.Time
	} else {
		maxUpdateDate = syncFallbackDate
	}

	return &maxUpdateDate, nil
}

func (t *TxnProcessing) syncProjects(ctx context.Context) ([]*dbmodels.Project, error) {
	// TODO: to be implemented

	return []*dbmodels.Project{}, nil
}
