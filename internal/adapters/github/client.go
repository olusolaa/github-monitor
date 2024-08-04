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
func NewClient(baseURL string, client *httpclient.Client) *Client {
	customClient := client
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

func (c *Client) GetCommits(ctx context.Context, owner, repoName, since, until string, commitsChan chan<- []Commit, errChan chan<- error) {
	defer close(commitsChan)
	defer close(errChan)

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

		select {
		case commitsChan <- *commits:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	out := &[]Commit{}
	if err := c.paginationManager.FetchAllPages(ctx, reqPath, params, processPage, out); err != nil {
		select {
		case errChan <- err:
		case <-ctx.Done():
		}
	}
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
