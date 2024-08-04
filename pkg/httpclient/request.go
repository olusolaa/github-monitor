package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	valid "github.com/asaskevich/govalidator"
	"github.com/google/go-querystring/query"
	"io"
	"net/http"
	"strings"
)

// RequestBuilder helps in building HTTP requests.
type RequestBuilder struct {
	baseURL string
	headers http.Header
}

// NewRequestBuilder creates a new instance of RequestBuilder.
func NewRequestBuilder(baseURL string) *RequestBuilder {
	return &RequestBuilder{
		baseURL: strings.TrimRight(baseURL, "/"),
		headers: make(http.Header),
	}
}

// SetHeader sets a header key-value pair for the request.
func (rb *RequestBuilder) SetHeader(key, value string) {
	rb.headers.Set(key, value)
}

// BuildRequest creates an HTTP request with the specified method, path, query parameters, and body.
func (rb *RequestBuilder) BuildRequest(ctx context.Context, method, path string, params interface{}, body interface{}) (*http.Request, error) {
	if (method == http.MethodGet || method == http.MethodDelete) && params != nil {
		_, err := valid.ValidateStruct(params)
		if err != nil {
			return nil, fmt.Errorf("invalid query parameters: %w", err)
		}

		v, _ := query.Values(params)
		path = path + "?" + v.Encode()
	}

	url := rb.baseURL + "/" + strings.TrimLeft(path, "/")

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set default content type if not already set
	if rb.headers.Get("Content-Type") == "" {
		rb.headers.Set("Content-Type", "application/json")
	}

	// Copy headers to the request
	for key, values := range rb.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return req, nil
}
