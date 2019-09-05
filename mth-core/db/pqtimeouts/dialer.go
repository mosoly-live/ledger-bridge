package pqtimeouts

import (
	"net"
	"time"
)

// TimeoutDialer is an alternative to pq's defaultDialer.
type TimeoutDialer struct {
	netDial        func(string, string) (net.Conn, error)                // Allow this to be stubbed for testing
	netDialTimeout func(string, string, time.Duration) (net.Conn, error) // Allow this to be stubbed for testing
	readTimeout    time.Duration
	writeTimeout   time.Duration
}

// NewTimeoutDialer creates a new dialer.
func NewTimeoutDialer(connectionString string) (d *TimeoutDialer, newConnectionString string, err error) {
	d = &TimeoutDialer{
		netDial:        net.Dial,
		netDialTimeout: net.DialTimeout,
	}
	d.readTimeout, d.writeTimeout, newConnectionString, err = parseConnectionString(connectionString)
	return
}

// Dial implements pq.Dialer.
func (t *TimeoutDialer) Dial(network string, address string) (net.Conn, error) {
	c, err := t.netDial(network, address)
	if err != nil {
		return nil, err
	}

	// If we don't have any timeouts set, just return a normal connection.
	if t.readTimeout == 0 && t.writeTimeout == 0 {
		return c, nil
	}

	// Otherwise we want a timeoutConn to handle the read and write deadlines for us.
	return &timeoutConn{conn: c, readTimeout: t.readTimeout, writeTimeout: t.writeTimeout}, nil
}

// DialTimeout implements pq.Dialer.
func (t *TimeoutDialer) DialTimeout(network string, address string, timeout time.Duration) (net.Conn, error) {
	c, err := t.netDialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}

	// If we don't have any timeouts set, just return a normal connection.
	if t.readTimeout == 0 && t.writeTimeout == 0 {
		return c, nil
	}

	// Otherwise we want a timeoutConn to handle the read and write deadlines for us.
	return &timeoutConn{conn: c, readTimeout: t.readTimeout, writeTimeout: t.writeTimeout}, nil
}
