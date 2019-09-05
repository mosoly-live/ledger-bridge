package kinesis

import (
	"encoding/json"
	"strings"

	"go.uber.org/zap"
)

// LogPutter puts all record to log
type LogPutter struct{}

// NewLogPutter creates new instance of LogPutter
func NewLogPutter() LogPutter {
	return LogPutter{}
}

// Put puts event asynchronously to stream. This method is thread-safe.
func (LogPutter) Put(e *Event) (err error) {
	if e == nil {
		return nil
	}

	var sb strings.Builder

	sb.WriteString("event: ")

	je := json.NewEncoder(&sb)
	err = je.Encode(e)
	if err != nil {
		return
	}

	zap.L().Info(sb.String())

	return
}
