package httpclient

import (
	"log"
	"net/http"
	"time"
)

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

func AuthMiddleware(token string) func(req *http.Request, next HTTPClient) (*http.Response, error) {
	return func(req *http.Request, next HTTPClient) (*http.Response, error) {
		req.Header.Set("Authorization", "Bearer "+token)
		return next.Do(req)
	}
}
