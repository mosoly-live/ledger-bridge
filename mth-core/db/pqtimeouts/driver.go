package pqtimeouts

import (
	"database/sql"
	"database/sql/driver"

	"github.com/lib/pq"
)

func init() {
	sql.Register("pq-timeouts", timeoutDriver{dialOpen: pq.DialOpen})
}

type timeoutDriver struct {
	dialOpen func(pq.Dialer, string) (driver.Conn, error) // Allow this to be stubbed for testing
}

func (t timeoutDriver) Open(connectionString string) (_ driver.Conn, err error) {
	dialer, newConnectionString, err := NewTimeoutDialer(connectionString)
	if err != nil {
		return nil, err
	}

	return t.dialOpen(dialer, newConnectionString)
}
