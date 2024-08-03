package httpclient

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of each request and response.
func LoggingMiddleware(req *http.Request, next HTTPClient) (*http.Response, error) {
	start := time.Now()
	log.Printf("Request: Method=%s, URL=%s, Headers=%v", req.Method, req.URL, req.Header)

	resp, err := next.Do(req)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Response Error: URL=%s, Duration=%s, Error=%v", req.URL, duration, err)
		return nil, err
	}
	log.Printf("Response: Status=%d, URL=%s, Duration=%s", resp.StatusCode, req.URL, duration)
	return resp, nil
}

// AuthMiddleware sets the Authorization header with the provided token.
func AuthMiddleware(token string) func(req *http.Request, next HTTPClient) (*http.Response, error) {
	return func(req *http.Request, next HTTPClient) (*http.Response, error) {
		req.Header.Set("Authorization", "Bearer "+token)
		return next.Do(req)
	}
}

// RetryMiddleware retries failed requests based on the provided retry count and backoff strategy.
func RetryMiddleware(retryCount int, initialBackoff time.Duration) func(req *http.Request, next HTTPClient) (*http.Response, error) {
	return func(req *http.Request, next HTTPClient) (*http.Response, error) {
		var resp *http.Response
		var err error
		backoff := initialBackoff

		for i := 0; i <= retryCount; i++ {
			resp, err = next.Do(req)
			if err == nil {
				if resp.StatusCode >= 200 && resp.StatusCode < 300 {
					return resp, nil
				}
			}

			if i < retryCount {
				time.Sleep(backoff)
				backoff *= 2 // Exponential backoff
			}
		}
		return resp, err
	}
}
