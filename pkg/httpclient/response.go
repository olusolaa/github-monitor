package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ResponseHandler decodes and processes HTTP responses.
type ResponseHandler struct{}

// NewResponseHandler creates a new instance of ResponseHandler.
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// HandleResponse decodes the HTTP response body into the provided interface.
// It also handles errors and extracts detailed error information if available.
func (rh *ResponseHandler) HandleResponse(resp *http.Response, out interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return rh.extractError(resp)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// extractError extracts detailed error information from the HTTP response.
func (rh *ResponseHandler) extractError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading error response body: %w", err)
	}

	errorType := "unexpected error"
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		errorType = "client error"
	} else if resp.StatusCode >= 500 {
		errorType = "server error"
	}

	return fmt.Errorf("%s: %s (status code: %d, body: %s)",
		errorType, http.StatusText(resp.StatusCode), resp.StatusCode, string(body))
}
