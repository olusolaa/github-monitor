package httpclient

import (
	"net/http"
)

// HTTPClient defines an interface for making HTTP requests, allowing for mocking and testing.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client represents a reusable HTTP client.
type Client struct {
	httpClient HTTPClient
	middleware []Middleware
}

// Middleware defines a function to process middleware.
type Middleware func(req *http.Request, next HTTPClient) (*http.Response, error)

// NewClient creates a new instance of Client with optional middleware.
func NewClient(httpClient HTTPClient, middleware ...Middleware) *Client {
	return &Client{
		httpClient: httpClient,
		middleware: middleware,
	}
}

// Do sends an HTTP request and returns an HTTP response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	finalClient := c.httpClient
	for i := len(c.middleware) - 1; i >= 0; i-- {
		mw := c.middleware[i]
		finalClient = &middlewareClient{client: finalClient, middleware: mw}
	}
	return finalClient.Do(req)
}

type middlewareClient struct {
	client     HTTPClient
	middleware Middleware
}

func (m *middlewareClient) Do(req *http.Request) (*http.Response, error) {
	return m.middleware(req, m.client)
}
