package services

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"time"

	"github.com/olusolaa/github-monitor/pkg/logger"
	"github.com/olusolaa/github-monitor/pkg/utils"
)

type MonitorService struct {
	repositoryService   RepositoryService
	commitService       CommitService
	gitHubService       GitHubService
	maxRetryAttempts    int
	initialRetryBackoff time.Duration
}

func NewMonitorService(repositoryService RepositoryService, commitService CommitService, githubService GitHubService, maxRetryAttempts int, initialRetryBackoff time.Duration) *MonitorService {
	return &MonitorService{
		repositoryService:   repositoryService,
		commitService:       commitService,
		gitHubService:       githubService,
		maxRetryAttempts:    maxRetryAttempts,
		initialRetryBackoff: initialRetryBackoff,
	}
}

// MonitorRepository oversees monitoring both repository and commit information for changes.
func (m *MonitorService) MonitorRepository(ctx context.Context, repositoryID int64) error {
	retryCount := 0
	for {
		err := m.syncRepositoryAndCommits(ctx, repositoryID)
		if err == nil {
			break
		}

		logger.LogError(err)
		retryCount++
		if retryCount >= m.maxRetryAttempts {
			return err
		}

		backoffDuration := utils.ExponentialBackoff(retryCount, m.initialRetryBackoff)
		time.Sleep(backoffDuration)
	}
	return nil
}

// syncRepositoryAndCommits fetches and updates both repository information and commits.
func (m *MonitorService) syncRepositoryAndCommits(ctx context.Context, repositoryID int64) error {
	if err := m.SyncRepositoryInfo(ctx, repositoryID); err != nil {
		return err
	}

	return m.MonitorRepositoryCommits(ctx, repositoryID)
}

func (m *MonitorService) MonitorRepositoryCommits(ctx context.Context, repositoryID int64) error {
	latestCommit, err := m.commitService.GetLatestCommit(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("could not get latest commit: %w", err)
	}

	var since string
	if latestCommit != nil {
		since = latestCommit.CommitDate.Format(time.RFC3339)
	}

	owner, name, err := m.repositoryService.GetOwnerAndRepoName(ctx, repositoryID)
	if err != nil {
		return fmt.Errorf("could not get repository owner and name: %w", err)
	}

	domainCommitsChan := make(chan []domain.Commit)
	errChan := make(chan error)

	go m.gitHubService.FetchCommits(ctx, owner, name, since, "", repositoryID, domainCommitsChan, errChan)

	for {
		select {
		case domainCommits, ok := <-domainCommitsChan:
			if !ok {
				return nil
			}
			if err := m.commitService.SaveCommits(ctx, domainCommits); err != nil {
				return fmt.Errorf("failed to save commits: %w", err)
			}
		case err := <-errChan:
			return fmt.Errorf("error fetching commits: %w", err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// SyncRepositoryInfo fetches and updates repository information.
func (m *MonitorService) SyncRepositoryInfo(ctx context.Context, repositoryID int64) error {
	owner, name, err := m.repositoryService.GetOwnerAndRepoName(ctx, repositoryID)
	if err != nil {
		return err
	}

	updatedRepository, err := m.gitHubService.FetchRepository(ctx, name, owner)
	if err != nil {
		return err
	}

	return m.repositoryService.UpsertRepository(ctx, updatedRepository)
}
