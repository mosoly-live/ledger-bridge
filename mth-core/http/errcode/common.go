package errcode

// Common error codes
const (
	CodeValidationError  Code = "VALIDATION_ERROR"   // 400
	CodeAuthRequired     Code = "AUTH_REQUIRED"      // 401
	CodeAuthTokenInvalid Code = "AUTH_TOKEN_INVALID" // 401
	CodeAccessDenied     Code = "ACCESS_DENIED"      // 403
	CodeResourceNotFound Code = "RESOURCE_NOT_FOUND" // 404
	CodeInternalError    Code = "INTERNAL_ERROR"     // 500
)
