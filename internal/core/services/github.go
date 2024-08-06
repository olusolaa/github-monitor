package services

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/internal/adapters/github"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type GitHubService interface {
	FetchRepository(ctx context.Context, owner, repoName string) (*domain.Repository, error)
	FetchCommits(ctx context.Context, owner, repoName, since, until string, repoID int64, commitsChan chan<- []domain.Commit, errChan chan<- error)
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
		logger.LogError(fmt.Errorf("failed to fetch repository info: %w", err))
		return nil, err
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
	logger.LogInfo(fmt.Sprintf("Repository fetched: %s/%s", owner, repoName))
	return repo, nil
}

func (s *gitHubService) FetchCommits(ctx context.Context, owner, repoName, since, until string, repoID int64, commitsChan chan<- []domain.Commit, errChan chan<- error) {
	apiCommitsChan := make(chan []github.Commit)
	apiErrChan := make(chan error)

	go s.client.GetCommits(ctx, owner, repoName, since, until, apiCommitsChan, apiErrChan)

	var encounteredError error

	defer func() {
		close(apiCommitsChan)
		close(apiErrChan)

		if encounteredError != nil {
			errChan <- encounteredError
		} else {
			errChan <- nil // Indicate completion without errors
		}
	}()

	for {
		select {
		case apiCommits, ok := <-apiCommitsChan:
			if !ok {
				logger.LogWarning("API commits channel closed unexpectedly")
			} else {
				domainCommits := s.convertToDomainCommits(apiCommits, repoID)
				select {
				case commitsChan <- domainCommits:
				case <-ctx.Done():
					encounteredError = ctx.Err()
					return
				}
			}
		case err, ok := <-apiErrChan:
			if err != nil {
				encounteredError = fmt.Errorf("error fetching commits: %w", err)
				logger.LogError(encounteredError)
				return
			} else if !ok {
				logger.LogWarning("API error channel closed unexpectedly")
			}
			return
		case <-ctx.Done():
			encounteredError = ctx.Err()
			logger.LogError(encounteredError)
			return
		}
	}
}

// convertToDomainCommits converts API commits to domain commits.
func (s *gitHubService) convertToDomainCommits(apiCommits []github.Commit, repoID int64) []domain.Commit {
	domainCommits := make([]domain.Commit, len(apiCommits))
	for i, commit := range apiCommits {
		domainCommits[i] = domain.Commit{
			RepositoryID: repoID,
			Hash:         commit.Sha,
			Message:      commit.Commit.Message,
			AuthorName:   commit.Commit.Committer.Name,
			AuthorEmail:  commit.Commit.Committer.Email,
			CommitDate:   commit.Commit.Committer.Date,
			URL:          commit.Commit.Url,
		}
	}
	return domainCommits
}
