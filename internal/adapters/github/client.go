package github

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"net/http"
)

// Client represents a GitHub API client.
type Client struct {
	httpClient        *httpclient.Client
	requestBuilder    *httpclient.RequestBuilder
	responseHandler   *httpclient.ResponseHandler
	paginationManager *PaginationManager
}

// NewClient creates a new GitHub API client with custom HTTP client settings.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	rateLimiter := NewGitHubRateLimiter()

	customClient := httpclient.NewClient(httpClient, rateLimiter.RateLimitMiddleware, httpclient.LoggingMiddleware, httpclient.AuthMiddleware("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"))
	requestBuilder := httpclient.NewRequestBuilder(baseURL)
	responseHandler := httpclient.NewResponseHandler()
	paginationManager := NewPaginationManager(requestBuilder, customClient, responseHandler)

	return &Client{
		httpClient:        customClient,
		requestBuilder:    requestBuilder,
		responseHandler:   responseHandler,
		paginationManager: paginationManager,
	}
}

// GetCommits fetches commits from a GitHub repository.
func (c *Client) GetCommits(ctx context.Context, owner, repoName, since, until string) ([]Commit, error) {
	var allCommits []Commit

	reqPath := fmt.Sprintf("/repos/%s/%s/commits", owner, repoName)
	params := CommitQueryParams{
		Since:   since,
		Until:   until,
		PerPage: 100,
	}

	processPage := func(data interface{}) error {
		commits, ok := data.(*[]Commit)
		if !ok {
			return fmt.Errorf("unexpected type %T for commit data", data)
		}
		allCommits = append(allCommits, *commits...)
		return nil
	}

	out := &[]Commit{}

	err := c.paginationManager.FetchAllPages(ctx, reqPath, params, processPage, out)
	return allCommits, err
}

// GetRepository fetches repository details from GitHub.
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)

	req, err := c.requestBuilder.BuildRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var repository Repository
	if err := c.responseHandler.HandleResponse(resp, &repository); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return &repository, nil
}
