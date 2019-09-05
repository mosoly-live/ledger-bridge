package kinesis

import "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/utils"

// EventType is a type for server event.
type EventType string

func (t EventType) String() string {
	return string(t)
}

// EventData holds server event data.
type EventData struct {
	Type        EventType   `json:"type"`
	Successful  bool        `json:"ok"`
	ContextData interface{} `json:"data,omitempty"`
}

// NewEventData creates an instance of EventData.
func NewEventData(t EventType) *EventData {
	return &EventData{Type: t}
}

// WithContextData assigns data and returns modified server event.
func (e *EventData) WithContextData(data interface{}) *EventData {
	if utils.IsNil(data) {
		e.ContextData = nil
	} else {
		e.ContextData = data
	}
	return e
}

// MarkAsSuccessful marks event as successful.
func (e *EventData) MarkAsSuccessful() *EventData {
	e.Successful = true
	return e
}
