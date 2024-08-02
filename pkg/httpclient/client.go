package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
)

const defaultTimeout = 60 * time.Second

// HTTPClient defines an interface for making HTTP requests, allowing for mocking and testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client represents a reusable HTTP client with a base URL and optional request hook.
type Client struct {
	httpClient HTTPClient
	baseURL    string
	debug      bool
	hook       RequestHook // Hook function to modify requests before sending
}

// NewClient creates a new instance of Client with the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    strings.TrimRight(baseURL, "/"),
		debug:      os.Getenv("ENV") != "production",
	}
}

// RequestHook defines a function signature for modifying requests.
type RequestHook func(req *http.Request) error

// SetHTTPClient allows setting a custom HTTP client (useful for testing).
func (cl *Client) SetHTTPClient(httpClient HTTPClient) {
	cl.httpClient = httpClient
}

// SetBaseURL sets a new base URL for the client.
func (cl *Client) SetBaseURL(baseURL string) {
	cl.baseURL = strings.TrimRight(baseURL, "/")
}

// SetRequestHook sets a hook function to modify requests before sending.
func (cl *Client) SetRequestHook(hook RequestHook) {
	cl.hook = hook
}

// Get performs a GET request with optional query parameters and unmarshals the response into the provided interface.
func (cl *Client) GetWithContext(ctx context.Context, path string, params interface{}, response interface{}) error {
	// Set up URL and parameters
	if params != nil {
		if _, err := valid.ValidateStruct(params); err != nil {
			return fmt.Errorf("invalid parameters: %w", err)
		}
		v, _ := query.Values(params)
		path = path + "?" + v.Encode()
	}

	url := cl.baseURL + "/" + strings.TrimLeft(path, "/")

	if cl.debug {
		log.Printf("httpclient: GET %s", url)
		log.Printf("httpclient: Request Params: %#v", params)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create GET request: %w", err)
	}

	return cl.doRequest(req, response)
}

// Post performs a POST request with optional body parameters and unmarshals the response into the provided interface.
func (cl *Client) PostWithContext(ctx context.Context, path string, params interface{}, response interface{}) error {
	url := cl.baseURL + "/" + strings.TrimLeft(path, "/")

	var body io.Reader
	if params != nil {
		if _, err := valid.ValidateStruct(params); err != nil {
			return fmt.Errorf("invalid parameters: %w", err)
		}

		data, err := json.Marshal(params)
		if err != nil {
			return fmt.Errorf("failed to marshal POST body: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	if cl.debug {
		log.Printf("httpclient: POST %s", url)
		log.Printf("httpclient: Request Params: %#v", params)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to create POST request: %w", err)
	}

	return cl.doRequest(req, response)
}

// doRequest executes the HTTP request and decodes the response.
func (cl *Client) doRequest(req *http.Request, response interface{}) error {
	req.Header.Set("Content-Type", "application/json")

	// Call the hook if it's set
	if cl.hook != nil {
		if err := cl.hook(req); err != nil {
			return fmt.Errorf("error in request hook: %w", err)
		}
	}

	resp, err := cl.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx response status: %d", resp.StatusCode)
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
