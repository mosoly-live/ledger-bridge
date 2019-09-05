package responses

// Error is error response from Client
type Error struct {
	Status  int64
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

// IsValidationErr checks if error is validation error
func (e Error) IsValidationErr() bool {
	return e.Code == "VALIDATION_ERROR"
}

// IsResourceNotFoundErr checks if error is resource not found error
func (e Error) IsResourceNotFoundErr() bool {
	return e.Code == "RESOURCE_NOT_FOUND"
}
