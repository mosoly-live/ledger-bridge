package txnprocessing

import (
	"time"
)

func getTxnProcessRunAt(now time.Time) (txnProcessRunAt time.Time) {
	txnProcessRunAt = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, time.UTC).Add(30 * time.Second)
	return
}
