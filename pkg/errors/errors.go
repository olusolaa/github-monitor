package errors

import (
	"errors"
	"github.com/go-chi/render"
	"net/http"
)

type errLogger interface {
	Error(args ...interface{})
}

// New returns *ErrResponse which implements Error interface
func New(errText string, statusCode ...int) *ErrResponse {
	sc := http.StatusBadRequest
	if len(statusCode) > 0 {
		sc = statusCode[0]
	}
	return &ErrResponse{
		Err:            errors.New(errText),
		HTTPStatusCode: sc,
		StatusText:     http.StatusText(sc),
		ErrorText:      errText,
	}
}

// CoverErr returns the err if it is an *ErrResponse and returns a
// defaultTo if otherwise. The value ErrResponse tells us that the error was handled
// and  should not notify sentry
func CoverErr(err, defaultTo error, logger errLogger) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*ErrResponse); ok {
		return err
	}
	logger.Error(err)
	return defaultTo
}

// ErrResponse renderer type for handling all sorts of errors.
type ErrResponse struct {
	Err            error  `json:"-"`                 // low-level runtime error
	HTTPStatusCode int    `json:"-"`                 // http response status code
	AppCode        int64  `json:"-"`                 // application-specific error code
	StatusText     string `json:"status"`            // users-level status message
	ErrorText      string `json:"message,omitempty"` // application-level error message, for debugging
}

func (e ErrResponse) Error() string {
	return e.Err.Error()
}

// Render sets the application-specific error code in AppCode.
func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

var (
	// ErrBadRequest returns status 400 Bad Request for malformed request body.
	ErrBadRequest = New(http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	// ErrNotFound returns status 404 Not Found for invalid resource request.
	ErrNotFound = New(http.StatusText(http.StatusNotFound), http.StatusNotFound)
	// ErrInternalServerError returns status 500 Internal Server Error.
	ErrInternalServerError = New(http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
)

func HandleError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch {
	//case errors.Is(err, services.ErrNotFound):
	//	status = http.StatusNotFound
	//	message = "Resource not found"
	//case errors.Is(err, services.ErrInvalidRequest):
	//	status = http.StatusBadRequest
	//	message = "Invalid request"
	default:
		status = http.StatusInternalServerError
		message = "Internal server error"
	}

	http.Error(w, message, status)
}
