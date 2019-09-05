package responder

import (
	"fmt"

	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/data/kinesis"
)

var (
	// EventPutter is the putter to send server events.
	EventPutter kinesis.EventPutter
	// PrincipalDataToEventFunc puts principal data into Kinesis event.
	PrincipalDataToEventFunc func(p interface{}, e *kinesis.Event)
	// EventSourceName defines the source name for server events.
	EventSourceName = "undefined-server-source"
)

type eventContext struct {
	Event     *kinesis.Event
	eventData *kinesis.EventData
}

// WithEvent sets up server event logging.
func (resp *Responder) WithEvent(typ kinesis.EventType, principal interface{}, contextData interface{}) *Responder {
	ec := &eventContext{}
	resp.eventContext = ec

	ec.Event = kinesis.NewEvent()
	ec.Event.Source = EventSourceName
	if principal != nil && PrincipalDataToEventFunc != nil {
		PrincipalDataToEventFunc(principal, ec.Event)
	}

	ec.eventData = kinesis.NewEventData(typ)
	ec.eventData.WithContextData(contextData)
	ec.Event.EventData = ec.eventData

	return resp
}

// WithContextData sets event context data.
func (resp *Responder) WithContextData(data interface{}) *Responder {
	resp.eventContext.eventData.WithContextData(data)
	return resp
}

func (resp *Responder) sendEvent() {
	if resp.eventContext == nil || resp.Event == nil {
		return
	}

	ec := resp.eventContext
	ev := ec.Event.WithHTTPRequest(resp.req)

	if resp.status < 400 {
		ec.eventData.MarkAsSuccessful()
	}

	if err := EventPutter.Put(ev); err != nil {
		resp.l.Error(fmt.Sprintf("failed to send event: %v", ev))
	}
}
