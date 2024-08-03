package services

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/internal/adapters/github"
	"github.com/olusolaa/github-monitor/internal/core/domain"
)

type GitHubService interface {
	FetchRepository(ctx context.Context, owner, repoName string) (*domain.Repository, error)
	FetchCommits(ctx context.Context, owner, repoName, since, until string, repoID int64) ([]domain.Commit, error)
}

type gitHubService struct {
	client *github.Client
}

func NewGitHubService(client *github.Client) GitHubService {
	return &gitHubService{client: client}
}

func (s *gitHubService) FetchRepository(ctx context.Context, owner, repoName string) (*domain.Repository, error) {
	apiRepo, err := s.client.GetRepository(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository info: %w", err)
	}

	repo := &domain.Repository{
		Owner:           owner,
		Name:            apiRepo.Name,
		Description:     apiRepo.Description,
		URL:             apiRepo.URL,
		Language:        apiRepo.Language,
		ForksCount:      apiRepo.ForksCount,
		StargazersCount: apiRepo.StargazersCount,
		OpenIssuesCount: apiRepo.OpenIssuesCount,
		WatchersCount:   apiRepo.WatchersCount,
		CreatedAt:       apiRepo.CreatedAt,
		UpdatedAt:       apiRepo.UpdatedAt,
	}
	return repo, nil
}

func (s *gitHubService) FetchCommits(ctx context.Context, owner, repoName, since, until string, repoID int64) ([]domain.Commit, error) {
	apiCommits, err := s.client.GetCommits(ctx, owner, repoName, since, until)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch commits from GitHub: %w", err)
	}

	commits := make([]domain.Commit, len(apiCommits))
	for i, commit := range apiCommits {
		commits[i] = domain.Commit{
			RepositoryID: repoID,
			Hash:         commit.Sha,
			Message:      commit.Commit.Message,
			AuthorName:   commit.Commit.Committer.Name,
			AuthorEmail:  commit.Commit.Committer.Email,
			CommitDate:   commit.Commit.Committer.Date,
			URL:          commit.Commit.Url,
		}
	}
	return commits, nil
}
