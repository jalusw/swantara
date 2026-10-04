package httpx

import "errors"

const (
	ErrBadRequestCode            = "ERR_BAD_REQUEST"
	ErrInternalServerCode        = "ERR_SERVICE"
	ErrUnauthorizedCode          = "ERR_UNAUTHORIZED"
	ErrTokenReusedCode           = "ERR_TOKEN_REUSED"
	ErrUnprocessableCode         = "ERR_UNPROCESSABLE"
	ErrNotFoundCode              = "ERR_NOT_FOUND"
	ErrForbiddenCode             = "ERR_FORBIDDEN"
	ErrServiceUnavailableCode    = "ERR_SERVICE_UNAVAILABLE"
	ErrValidationCode            = "ERR_VALIDATION"
	ErrConflictCode              = "ERR_CONFLICT"
	ErrRateLimitCode             = "ERR_RATE_LIMIT"
	ErrUnsupportedMediaTypeCode  = "ERR_UNSUPPORTED_MEDIA_TYPE"
	DefaultSuccessMessage        = "Request processed successfully."
	DefaultErrorMessage          = "Something went wrong. Please try again."
	DefaultUnauthorizedMessage   = "Unauthorized. Please sign in to continue."
	DefaultInternalServerMessage = "Internal server error."
	DefaultUnprocessableMessage  = "Unable to process the request. Please check your input."
	DefaultNotFoundMessage       = "The requested resource was not found."
	DefaultBadRequestMessage     = "Bad request. Please check your input."
)

var (
	ErrMissingAuthorizationHeader   = errors.New("authorization header is not provided")
	ErrMissingOrganizationHeader    = errors.New("x-organization-id header is not provided")
	ErrMissingOrganizationInContext = errors.New("organization is not found in context")
	ErrMissingUserInContext         = errors.New("user is not found in context")
	ErrSessionIDRequired            = errors.New("session id is required")
)
