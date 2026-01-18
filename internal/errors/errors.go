// Package errors provides comprehensive error handling for the application.
package errors

import (
	"fmt"
	"io"
	"os"
)

// ErrorType represents the type of error.
type ErrorType int

const (
	// Auth errors (1xx)
	ErrAuth ErrorType = iota + 100
	ErrAuthExpired
	ErrAuthRevoked
	ErrAuthOffline

	// API errors (2xx)
	ErrAPI
	ErrAPIRateLimit
	ErrAPINotFound
	ErrAPIForbidden
	ErrAPIServerError
	ErrAPINetwork

	// Cache errors (3xx)
	ErrCache
	ErrCacheRead
	ErrCacheWrite
	ErrCacheExpired

	// Validation errors (4xx)
	ErrValidation
	ErrInvalidInput

	// System errors (5xx)
	ErrSystem
	ErrIO
	ErrConfig
)

// Error represents an application error with context.
type Error struct {
	Type           ErrorType
	Message        string
	Original       error
	UserSuggestion string
	Recoverable    bool
}

// New creates a new Error.
func New(errType ErrorType, message string) *Error {
	return &Error{
		Type:        errType,
		Message:     message,
		Recoverable: true,
	}
}

// Newf creates a new Error with format.
func Newf(errType ErrorType, format string, args ...interface{}) *Error {
	return &Error{
		Type:        errType,
		Message:     fmt.Sprintf(format, args...),
		Recoverable: true,
	}
}

// Wrap wraps an existing error.
func Wrap(err error, errType ErrorType, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Type:        errType,
		Message:     message,
		Original:    err,
		Recoverable: true,
	}
}

// Wrapf wraps an existing error with format.
func Wrapf(err error, errType ErrorType, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Type:        errType,
		Message:     fmt.Sprintf(format, args...),
		Original:    err,
		Recoverable: true,
	}
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Original)
	}
	return e.Message
}

// Is checks if the error is of a specific type.
func (e *Error) Is(errType ErrorType) bool {
	return e.Type == errType
}

// WithSuggestion adds a suggestion to the error.
func (e *Error) WithSuggestion(suggestion string) *Error {
	e.UserSuggestion = suggestion
	return e
}

// NotRecoverable marks the error as not recoverable.
func (e *Error) NotRecoverable() *Error {
	e.Recoverable = false
	return e
}

// UserMessage returns a user-friendly message.
func (e *Error) UserMessage() string {
	switch e.Type {
	case ErrAuth, ErrAuthExpired, ErrAuthRevoked:
		return "Authentication required. Please log in again."
	case ErrAuthOffline:
		return "You appear to be offline. Some features may not be available."
	case ErrAPIRateLimit:
		return "Too many requests. Please wait a moment and try again."
	case ErrAPINotFound:
		return "The requested item was not found."
	case ErrAPIForbidden:
		return "You don't have permission to access this resource."
	case ErrAPIServerError:
		return "A server error occurred. Please try again later."
	case ErrAPINetwork:
		return "Network error. Please check your connection."
	case ErrCacheRead, ErrCacheWrite:
		return "There was a problem accessing cached data."
	case ErrCacheExpired:
		return "Cached data has expired. Refreshing..."
	case ErrValidation, ErrInvalidInput:
		return "Invalid input. Please check your entry."
	case ErrIO:
		return "A file system error occurred."
	case ErrConfig:
		return "Configuration error. Please check your settings."
	default:
		return e.Message
	}
}

// GetSuggestion returns a suggestion for the user.
func (e *Error) GetSuggestion() string {
	if e.UserSuggestion != "" {
		return e.UserSuggestion
	}

	switch e.Type {
	case ErrAuthExpired, ErrAuthRevoked:
		return "Run 'google-classroom auth login' to re-authenticate."
	case ErrAPIRateLimit:
		return "Wait a few seconds before retrying."
	case ErrAPINetwork:
		return "Check your internet connection."
	case ErrCacheExpired:
		return "Press 'r' to refresh the data."
	default:
		return "Press 'r' to retry or 'q' to quit."
	}
}

// IsRateLimitError checks if the error is a rate limit error.
func IsRateLimitError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrAPIRateLimit
	}
	return false
}

// IsAuthError checks if the error is an authentication error.
func IsAuthError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrAuth || e.Type == ErrAuthExpired || e.Type == ErrAuthRevoked
	}
	return false
}

// IsNotFoundError checks if the error is a not found error.
func IsNotFoundError(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Type == ErrAPINotFound
	}
	return false
}

// IsRecoverable checks if the error is recoverable.
func IsRecoverable(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Recoverable
	}
	return true
}

// Handler provides error handling functionality.
type Handler struct {
	onError func(*Error)
	onRetry func(*Error) bool
}

// NewHandler creates a new error handler.
func NewHandler() *Handler {
	return &Handler{
		onError: defaultOnError,
		onRetry: defaultOnRetry,
	}
}

// Handle handles an error.
func (h *Handler) Handle(err error) *Error {
	if err == nil {
		return nil
	}

	// If already an Error type, handle it
	if e, ok := err.(*Error); ok {
		h.onError(e)
		return e
	}

	// Wrap other errors
	e := Wrap(err, ErrAPI, "An error occurred")
	h.onError(e)
	return e
}

// SetOnError sets the error callback.
func (h *Handler) SetOnError(fn func(*Error)) {
	h.onError = fn
}

// SetOnRetry sets the retry callback.
func (h *Handler) SetOnRetry(fn func(*Error) bool) {
	h.onRetry = fn
}

// defaultOnError is the default error callback.
func defaultOnError(e *Error) {
	// Log the error
	fmt.Fprintf(stderr, "Error: %s\n", e.UserMessage())
}

// defaultOnRetry is the default retry callback.
func defaultOnRetry(e *Error) bool {
	return e.Recoverable
}

// stderr is used for error output.
var stderr io.Writer = os.Stderr

// SetStderr sets the stderr writer.
func SetStderr(w io.Writer) {
	stderr = w
}
