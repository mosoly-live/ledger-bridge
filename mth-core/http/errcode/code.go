package errcode

// Code is a typed representation of HTTP error codes.
type Code string

// WithMessage creates error response from error code and message.
func (c Code) WithMessage(message string) *Response {
	return &Response{
		Code:    string(c),
		Message: message,
	}
}

// WithError creates error response from error code and error.
func (c Code) WithError(err error) *Response {
	return c.WithMessage(err.Error())
}

// Response is generic error response for HTTP requests.
type Response struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
