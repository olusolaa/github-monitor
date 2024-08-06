package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// Severity levels for errors
type Severity string

const (
	Critical Severity = "Critical"
	Warning  Severity = "Warning"
	Info     Severity = "Info"
)

// CustomError defines a structure for custom errors with additional context
type CustomError struct {
	Code       string
	Message    string
	Err        error
	Severity   Severity
	Timestamp  time.Time
	StackTrace string
}

// Error returns the error message string
func (e *CustomError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s][%s] %s: %v", e.Timestamp.Format(time.RFC3339), e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s][%s] %s", e.Timestamp.Format(time.RFC3339), e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *CustomError) Unwrap() error {
	return e.Err
}

// New creates a new custom error with stack trace
func New(code, message string, err error, severity Severity) error {
	return &CustomError{
		Code:       code,
		Message:    message,
		Err:        err,
		Severity:   severity,
		Timestamp:  time.Now(),
		StackTrace: getStackTrace(),
	}
}

// getStackTrace captures the current stack trace
func getStackTrace() string {
	stackBuf := make([]byte, 1024)
	stackBuf = stackBuf[:runtime.Stack(stackBuf, false)]
	return string(stackBuf)
}

// APIError defines a structure for API error responses
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// HandleError handles errors and sends an appropriate HTTP response
func HandleError(w http.ResponseWriter, err error) {
	var apiErr APIError

	var e *CustomError
	switch {
	case errors.As(err, &e):
		apiErr = APIError{
			Code:    mapSeverityToHTTPStatus(e.Severity),
			Message: e.Message,
			Details: e.Err.Error(),
		}
	default:
		apiErr = APIError{
			Code:    http.StatusInternalServerError,
			Message: "Internal Server Error",
			Details: err.Error(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErr.Code)
	json.NewEncoder(w).Encode(apiErr)
}

// mapSeverityToHTTPStatus maps custom error severity to HTTP status codes
func mapSeverityToHTTPStatus(severity Severity) int {
	switch severity {
	case Critical:
		return http.StatusInternalServerError
	case Warning:
		return http.StatusBadRequest
	case Info:
		return http.StatusOK
	default:
		return http.StatusInternalServerError
	}
}
