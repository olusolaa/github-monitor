package github

import (
	"github.com/olusolaa/github-monitor/pkg/errors"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter manages rate limiting based on GitHub's rate limit headers.
type RateLimiter struct {
	mu        sync.Mutex
	remaining int
	reset     time.Time
}

// NewGitHubRateLimiter initializes a new GitHubRateLimiter.
func NewGitHubRateLimiter() *RateLimiter {
	return &RateLimiter{}
}

// RateLimitMiddleware controls the rate of requests based on GitHub's rate limit headers.
func (rl *RateLimiter) RateLimitMiddleware(req *http.Request, next httpclient.HTTPClient) (*http.Response, error) {
	rl.mu.Lock()
	now := time.Now()
	if rl.remaining == 0 && rl.reset.After(now) {
		rl.mu.Unlock()
		timeUntilReset := time.Until(rl.reset)
		return nil, &RateLimitExceededError{ResetTime: rl.reset, RetryAfter: timeUntilReset}
	}
	rl.mu.Unlock()

	resp, err := next.Do(req)
	if err != nil {
		return nil, errors.New("HTTP_REQUEST_ERROR", "failed to execute request", err, errors.Critical)
	}

	rl.updateRateLimit(resp)
	return resp, nil
}

// updateRateLimit updates the rate limit state based on the GitHub API response headers.
func (rl *RateLimiter) updateRateLimit(resp *http.Response) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if r, err := strconv.Atoi(remaining); err == nil {
			rl.remaining = r
		}
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		if r, err := strconv.ParseInt(reset, 10, 64); err == nil {
			rl.reset = time.Unix(r, 0)
		}
	}
}

// RateLimitExceededError represents an error when the rate limit is exceeded.
type RateLimitExceededError struct {
	ResetTime  time.Time
	RetryAfter time.Duration
}

func (e *RateLimitExceededError) Error() string {
	return "rate limit exceeded, retry after " + e.RetryAfter.String()
}
