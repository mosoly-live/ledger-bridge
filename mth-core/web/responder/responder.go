package responder

import (
	"net/http"
	"reflect"

	"github.com/go-openapi/runtime"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/http/errcode"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/log"
	webcontext "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/web/context"
)

// Responder implements middleware.Responder, to use while generating responses in our handlers.
// It mimics what each piece of generated responder code does in our microservices, and plus has
// some logic to log problems and events automatically and create responses easily.
//
// The general practice when we write our error responses is:
//   - BadRequest: Custom error code with message
//   - Forbidden: Custom error code with message (status = 403)
//   - AccessDenied: Common error code and message (status = 403)
//   - NotFound: Common error code and message
//   - InternalError: Common error code and message
//
// For all these reasons, we have easy responses which are like:
//   return resp.InternalError(err, "this is the log message (reason)").Msg("oops! sorry, request failed")
//
// So it's possible to override any field. However, it's also possible to construct responses from scratch with ease:
//   return resp.Status(400).Code(errcode.CodeBadNumber).Msg("the number is too large").Err(moreSpecificErr)
//
// This will cause "BAD_NUMBER" and "the number is too large" in response, while it logs the error with same
// message. If we need a more specific logging message. Then we can provide it with Reason().
type Responder struct {
	req *http.Request
	l   *log.Logger

	status int
	err    error
	code   errcode.Code
	msg    string
	reason string
	body   interface{}

	*eventContext
}

// New creates a new responder.
func New(req *http.Request) (resp *Responder) {
	resp = &Responder{req: req}

	// Set up logger with correlation ID if it's existing. Otherwise, use a logger without correlation ID.
	if cID := webcontext.CorrelationID(req.Context()); len(cID) > 0 {
		resp.l = log.With(log.CorrelationID(cID))
	} else {
		resp.l = log.With()
	}

	return resp
}

// L returns logger.
func (resp *Responder) L() *log.Logger {
	return resp.l
}

// WriteResponse writes HTTP response using prepared data.
func (resp *Responder) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	if resp.status == 0 {
		resp.status = 200
	}

	resp.prepareDefaultMessages()
	resp.trySetErrorResponse()

	// If there is a reason (log message) or an error, do the logging.
	if len(resp.reason) > 0 || resp.err != nil {
		if resp.err != nil {
			resp.l.Warn(resp.reason, log.Err(resp.err))
		} else {
			resp.l.Warn(resp.reason)
		}
	}

	resp.sendEvent()

	resp.write(rw, producer)
}

func (resp *Responder) prepareDefaultMessages() {
	// If status code is fine, do not prepare any messages.
	if resp.status < 400 {
		return
	}

	if len(resp.msg) == 0 {
		resp.msg = "request failed"
	}
	if len(resp.reason) == 0 {
		resp.reason = resp.msg
	}
}

// trySetErrorResponse tries to set error response. If the response has a failure response,
// body is not yet set and there is a specified error code, then body is prepared with error response.
func (resp *Responder) trySetErrorResponse() {
	if resp.body == nil && len(resp.code) > 0 && resp.status >= 400 {
		resp.body = resp.code.WithMessage(resp.msg)
	}
}

// write writes response:
//   1) Set header fields.
//   2) Write header with status.
//   3) Write body.
func (resp *Responder) write(rw http.ResponseWriter, producer runtime.Producer) {
	// If response is nil, remove Content-Type header field only.
	if resp.body == nil {
		rw.Header().Del(runtime.HeaderContentType)
	}

	rw.WriteHeader(resp.status)

	if resp.body != nil {
		resp.fixZeroBody() // Try not to marshal to "null" if body is not nil
		if err := producer.Produce(rw, resp.body); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

func (resp *Responder) fixZeroBody() {
	v := reflect.ValueOf(resp.body)
	if v.Kind() == reflect.Slice && v.Len() == 0 {
		resp.body = reflect.MakeSlice(v.Type(), 0, 50).Interface()
	}
}
