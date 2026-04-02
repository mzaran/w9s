// Package errs provides typed errors for the w9s application.
package errs

import "fmt"

// ErrorCode identifies the category of a W9sError.
type ErrorCode int

const (
	// ErrConnection indicates a failure to connect to the Warewulf server.
	ErrConnection ErrorCode = iota
	// ErrAuth indicates invalid credentials.
	ErrAuth
	// ErrForbidden indicates insufficient permissions.
	ErrForbidden
	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound
	// ErrTimeout indicates a request timeout.
	ErrTimeout
	// ErrAPIDisabled indicates the Warewulf API is not enabled.
	ErrAPIDisabled
	// ErrValidation indicates invalid input or configuration.
	ErrValidation
	// ErrPowerUnavailable indicates IPMI/power management is not available.
	ErrPowerUnavailable
)

// W9sError is the standard error type for the w9s application.
type W9sError struct {
	Code    ErrorCode
	Message string
	Err     error
}

// Error implements the error interface.
func (e *W9sError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *W9sError) Unwrap() error {
	return e.Err
}

// Wrap creates a W9sError wrapping an existing error.
func Wrap(err error, code ErrorCode, msg string) *W9sError {
	return &W9sError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// NewConnectionError creates a connection error.
func NewConnectionError(err error, msg string) *W9sError {
	return Wrap(err, ErrConnection, msg)
}

// NewAuthError creates an authentication error.
func NewAuthError(err error, msg string) *W9sError {
	return Wrap(err, ErrAuth, msg)
}

// NewForbiddenError creates a forbidden/permissions error.
func NewForbiddenError(err error, msg string) *W9sError {
	return Wrap(err, ErrForbidden, msg)
}

// NewNotFoundError creates a not-found error.
func NewNotFoundError(err error, msg string) *W9sError {
	return Wrap(err, ErrNotFound, msg)
}

// NewTimeoutError creates a timeout error.
func NewTimeoutError(err error, msg string) *W9sError {
	return Wrap(err, ErrTimeout, msg)
}

// NewAPIDisabledError creates an API-disabled error.
func NewAPIDisabledError(err error, msg string) *W9sError {
	return Wrap(err, ErrAPIDisabled, msg)
}

// NewValidationError creates a validation error.
func NewValidationError(err error, msg string) *W9sError {
	return Wrap(err, ErrValidation, msg)
}

// NewPowerUnavailableError creates a power-unavailable error.
func NewPowerUnavailableError(err error, msg string) *W9sError {
	return Wrap(err, ErrPowerUnavailable, msg)
}
