package responder

import "gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/http/errcode"

// Status sets response status code.
func (resp *Responder) Status(status int) *Responder {
	resp.status = status
	return resp
}

// Code sets error response code.
func (resp *Responder) Code(code errcode.Code) *Responder {
	resp.code = code
	return resp
}

// Msg sets error response message.
func (resp *Responder) Msg(msg string) *Responder {
	resp.msg = msg
	return resp
}

// Err sets error for logging.
func (resp *Responder) Err(err error) *Responder {
	resp.err = err
	return resp
}

// Reason sets error reason (a more descriptive error message) for logging.
func (resp *Responder) Reason(reason string) *Responder {
	resp.reason = reason
	return resp
}

// Body sets response body to respond with.
func (resp *Responder) Body(body interface{}) *Responder {
	resp.body = body
	return resp
}
