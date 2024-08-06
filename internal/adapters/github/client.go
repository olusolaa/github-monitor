package github

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/pkg/errors"
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
	reqPath := fmt.Sprintf("/repos/%s/%s/commits", owner, repoName)
	params := CommitQueryParams{
		Since:   since,
		Until:   until,
		PerPage: 100,
	}

	processPage := func(data interface{}) error {
		commits, ok := data.(*[]Commit)
		if !ok {
			return errors.New("PROCESS_PAGE_ERROR", "unexpected type for commit data", fmt.Errorf("unexpected type %T", data), errors.Critical)
		}

		select {
		case commitsChan <- *commits:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	}

	out := &[]Commit{}
	fetchErr := c.paginationManager.FetchAllPages(ctx, reqPath, params, processPage, out)

	// Signal completion
	select {
	case errChan <- fetchErr:
	case <-ctx.Done():
		// Handle context cancellation, don't block
	}
}

// GetRepository fetches repository details from GitHub.
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)

	req, err := c.requestBuilder.BuildRequest(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, errors.New("BUILD_REQUEST_ERROR", "failed to build request", err, errors.Critical)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.New("EXECUTE_REQUEST_ERROR", "failed to execute request", err, errors.Critical)
	}
	defer resp.Body.Close()

	var repository Repository
	if err := c.responseHandler.HandleResponse(resp, &repository); err != nil {
		return nil, errors.New("HANDLE_RESPONSE_ERROR", "failed to handle response", err, errors.Critical)
	}

	return &repository, nil
}
