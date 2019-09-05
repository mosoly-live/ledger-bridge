package pqtimeouts

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

func parseConnectionString(connection string) (readTimeout time.Duration, writeTimeout time.Duration,
	newConnectionString string, err error) {

	// Look for read_timeout and write_timeout in the connection string and extract the values.
	// read_timeout and write_timeout need to be removed from the connection string before calling pq as well.
	var newConnectionSettings []string

	// If the connection is specified as a URL, use the parsing function in lib/pq to turn it into options.
	if strings.HasPrefix(connection, "postgres://") || strings.HasPrefix(connection, "postgresql://") {
		connection, err = pq.ParseURL(connection)
		if err != nil {
			return
		}
	}

	for _, setting := range strings.Fields(connection) {
		s := strings.Split(setting, "=")

		switch s[0] {
		case "read_timeout":
			val, err := strconv.Atoi(s[1])
			if err != nil {
				return 0, 0, "", fmt.Errorf("error interpreting value for read_timeout: %v", err)
			}
			readTimeout = time.Duration(val) * time.Millisecond // timeout is in milliseconds

		case "write_timeout":
			val, err := strconv.Atoi(s[1])
			if err != nil {
				return 0, 0, "", fmt.Errorf("error interpreting value for write_timeout: %v", err)
			}
			writeTimeout = time.Duration(val) * time.Millisecond // timeout is in milliseconds

		default:
			newConnectionSettings = append(newConnectionSettings, setting)
		}
	}

	newConnectionString = strings.Join(newConnectionSettings, " ")

	return
}
