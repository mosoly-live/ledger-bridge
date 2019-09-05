package kinesis

import (
	"net/http"
	"time"

	"github.com/go-openapi/strfmt"
)

// HTTPData contains data from HTTP request
type HTTPData struct {
	ClientIP        string `json:"ip,omitempty"`
	ClientIPCountry string `json:"co,omitempty"`
	UserAgent       string `json:"ua,omitempty"`
}

// Event is generic structure for Kinesis event
type Event struct {
	Source        string      `json:"src"`
	TimestampNano int64       `json:"tsn,string"`
	IsTest        bool        `json:"tst,omitempty"`
	UserID        int64       `json:"uid,omitempty"`
	UserUUID      string      `json:"uuid,omitempty"`
	MerchantID    int64       `json:"mid,omitempty"`
	UserName      string      `json:"scuid,omitempty"`
	HTTPData      *HTTPData   `json:"http,omitempty"`
	EventData     interface{} `json:"ed,omitempty"`
}

// NewEvent creates an instance of event
func NewEvent() *Event {
	return &Event{TimestampNano: time.Now().UTC().UnixNano()}
}

// WithHTTPRequest enriches event with the HTTP request data (like client IP address)
func (e *Event) WithHTTPRequest(r *http.Request) *Event {
	if r == nil {
		e.HTTPData = nil
		return e
	}

	headers := r.Header

	var clientIP string
	if clientIP = headers.Get("Cf-Connecting-Ip"); len(clientIP) == 0 {
		clientIP = r.RemoteAddr
	}

	e.HTTPData = &HTTPData{
		ClientIP:        clientIP,
		ClientIPCountry: headers.Get("Cf-Ipcountry"),
		UserAgent:       headers.Get("User-Agent"),
	}
	return e
}

// WithUserIDs assigns User ID and UUID.
func (e *Event) WithUserIDs(userID int64, userUUID strfmt.UUID) *Event {
	e.UserID = userID
	e.UserUUID = string(userUUID)
	return e
}

// WithIsTest assigns isTest.
func (e *Event) WithIsTest(isTest bool) *Event {
	e.IsTest = isTest
	return e
}

// EventPutter is an interface to put events into some processing pipeline
type EventPutter interface {
	Put(e *Event) (err error)
}
