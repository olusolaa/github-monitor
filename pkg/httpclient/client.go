package httpclient

import (
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient HTTPClient
	middleware []Middleware
}

type Middleware func(req *http.Request, next HTTPClient) (*http.Response, error)

func NewClient(httpClient HTTPClient, middleware ...Middleware) *Client {
	return &Client{
		httpClient: httpClient,
		middleware: middleware,
	}
}

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
