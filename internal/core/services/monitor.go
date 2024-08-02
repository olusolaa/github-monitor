package services

import (
	"context"
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
	// Synchronize repository information
	if err := m.SyncRepositoryInfo(ctx, repositoryID); err != nil {
		return err
	}

	// Monitor and save new commits
	return m.MonitorRepositoryCommits(ctx, repositoryID)
}

// MonitorRepositoryCommits fetches new commits and updates the database.
func (m *MonitorService) MonitorRepositoryCommits(ctx context.Context, repositoryID int64) error {
	latestCommit, err := m.commitService.GetLatestCommit(ctx, repositoryID)
	if err != nil {
		return err
	}

	var since string
	if latestCommit != nil {
		since = latestCommit.CommitDate.Format(time.RFC3339)
	}

	owner, name, err := m.repositoryService.GetOwnerAndRepoName(ctx, repositoryID)
	if err != nil {
		return err
	}

	commits, err := m.gitHubService.FetchCommits(ctx, owner, name, since, "", repositoryID)
	if err != nil {
		return err
	}

	return m.commitService.SaveCommits(ctx, commits)
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
