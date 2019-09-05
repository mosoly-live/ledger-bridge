package responder

import (
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/http/errcode"
)

// OK sets response to be an "OK" response with no errors.
func (resp *Responder) OK(body interface{}) *Responder {
	resp.status = 200
	resp.body = body
	return resp
}

// Created sets response to be a "created" response with no errors.
func (resp *Responder) Created(body interface{}) *Responder {
	resp.status = 201
	resp.body = body
	return resp
}

// NoContent sets response to be a "no content" response with no errors.
func (resp *Responder) NoContent() *Responder {
	resp.status = 204
	return resp
}

// BadRequest sets response to be a "bad request" response.
func (resp *Responder) BadRequest(code errcode.Code, msg string, reason ...string) *Responder {
	return resp.prepare(400, nil, code, msg, reason...)
}

// ValidationError sets response to be a "validation error" response.
func (resp *Responder) ValidationError(msg string) *Responder {
	return resp.prepare(400, nil, errcode.CodeValidationError, msg, msg)
}

// Forbidden sets response to be a "forbidden" response.
func (resp *Responder) Forbidden(code errcode.Code, msg string, reason ...string) *Responder {
	return resp.prepare(403, nil, code, msg, reason...)
}

// AccessDenied sets response to be an "access denied" response.
func (resp *Responder) AccessDenied(reason ...string) *Responder {
	return resp.prepare(403, nil, errcode.CodeAccessDenied, "access denied to resource", reason...)
}

// NotFound sets response to be a "not found" response.
func (resp *Responder) NotFound(err error, reason ...string) *Responder {
	return resp.prepare(404, err, errcode.CodeResourceNotFound, "resource not found", reason...)
}

// InternalError sets response to be an "internal error" response.
func (resp *Responder) InternalError(err error, reason ...string) *Responder {
	return resp.prepare(500, err, errcode.CodeInternalError, "something went wrong", reason...)
}

func (resp *Responder) prepare(status int, err error, code errcode.Code, msg string, reason ...string) *Responder {
	resp.Status(status).Code(code).Msg(msg).Err(err)
	if reason != nil {
		resp.Reason(reason[0])
	}
	return resp
}
