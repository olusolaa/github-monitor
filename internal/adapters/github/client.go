package github

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"net/http"
)

const githubAPIBaseURL = "https://api.github.com"

// Client represents a GitHub API client
type Client struct {
	httpClient *httpclient.Client
	token      string
}

// NewClient creates a new GitHub API client
func NewClient(token string) *Client {
	client := &Client{
		httpClient: httpclient.NewClient(githubAPIBaseURL),
		token:      token,
	}

	// Set the request hook to add the Authorization header
	client.httpClient.SetRequestHook(func(req *http.Request) error {
		req.Header.Set("Authorization", "token "+token)
		return nil
	})

	return client
}

// CommitQueryParams holds the parameters for querying commits
type CommitQueryParams struct {
	Since string `url:"since,omitempty"`
	Until string `url:"until,omitempty"`
}

// GetRepository fetches repository details from GitHub
func (c *Client) GetRepository(ctx context.Context, owner, repo string) (*domain.Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	var repository domain.Repository

	if err := c.httpClient.GetWithContext(ctx, path, nil, &repository); err != nil {
		return nil, fmt.Errorf("failed to get repository details: %w", err)
	}
	return &repository, nil
}

// GetCommits fetches commits from a GitHub repository
func (c *Client) GetCommits(ctx context.Context, owner, repoName, since, until string) ([]Commit, error) {
	path := fmt.Sprintf("/repos/%s/%s/commits", owner, repoName)
	params := CommitQueryParams{
		Since: since,
		Until: until,
	}

	var commits []Commit
	if err := c.httpClient.GetWithContext(ctx, path, params, &commits); err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	return commits, nil
}
