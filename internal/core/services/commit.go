package services

import (
	"context"
	"github.com/olusolaa/github-monitor/pkg/pagination"
	"time"

	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type CommitService interface {
	FetchCommits(ctx context.Context, repoID int64, startTime, endTime string) ([]domain.Commit, error)
	SaveCommits(ctx context.Context, commits []domain.Commit) error
	GetLatestCommit(ctx context.Context, repoID int64) (*domain.Commit, error)
	GetCommitsByRepository(ctx context.Context, repoID int64, page, pageSize int) ([]domain.Commit, *pagination.Pagination, error)
	ResetCollection(ctx context.Context, repoID int64, startTime time.Time) error
	GetTopCommitAuthors(ctx context.Context, repoID int64, limit int) ([]domain.CommitAuthor, error)
}

type commitService struct {
	ghService         GitHubService
	repositoryService RepositoryService
	commitRepo        *postgresdb.CommitRepository
}

func NewCommitService(ghService GitHubService, repositoryService RepositoryService, commitRepo *postgresdb.CommitRepository) CommitService {
	return &commitService{ghService: ghService, repositoryService: repositoryService, commitRepo: commitRepo}
}

// FetchCommits fetches commits for a given repository and time range from GitHub
func (s *commitService) FetchCommits(ctx context.Context, repoID int64, startTime, endTime string) ([]domain.Commit, error) {
	owner, repoName, err := s.repositoryService.GetOwnerAndRepoName(ctx, repoID)
	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	commits, err := s.ghService.FetchCommits(ctx, owner, repoName, startTime, endTime, repoID)
	if err != nil {
		logger.LogError(err)
		return nil, err
	}
	return commits, nil
}

// SaveCommits saves the provided commits into the repository
func (s *commitService) SaveCommits(ctx context.Context, commits []domain.Commit) error {
	if err := s.commitRepo.Save(ctx, commits); err != nil {
		logger.LogError(err)
		return err
	}
	return nil
}

// GetLatestCommit retrieves the most recent commit for a given repository
func (s *commitService) GetLatestCommit(ctx context.Context, repoID int64) (*domain.Commit, error) {
	latestCommit, err := s.commitRepo.GetLatestCommitByRepositoryID(ctx, repoID)
	if err != nil {
		logger.LogError(err)
		return nil, err
	}
	return latestCommit, nil
}

func (s *commitService) GetCommitsByRepository(ctx context.Context, repoID int64, page, pageSize int) ([]domain.Commit, *pagination.Pagination, error) {
	commits, totalItems, err := s.commitRepo.GetCommitsByRepositoryID(ctx, repoID, page, pageSize)
	if err != nil {
		logger.LogError(err)
		return nil, nil, err
	}

	pg := pagination.NewPagination(page, pageSize, totalItems)
	return commits, pg, nil
}

func (s *commitService) GetTopCommitAuthors(ctx context.Context, repoID int64, limit int) ([]domain.CommitAuthor, error) {
	return s.commitRepo.GetTopCommitAuthors(ctx, repoID, limit)
}

func (s *commitService) ResetCollection(ctx context.Context, repoID int64, startTime time.Time) error {
	// Start transaction to ensure atomicity
	tx, err := s.commitRepo.BeginTx(ctx)
	if err != nil {
		logger.LogError(err)
		return err
	}

	// Delete existing commits
	if err := s.commitRepo.DeleteCommitsByRepositoryID(ctx, repoID); err != nil {
		logger.LogError(err)
		tx.Rollback()
		return err
	}

	// Fetch new commits from the start time
	startTimeStr := startTime.Format(time.RFC3339)
	commits, err := s.FetchCommits(ctx, repoID, startTimeStr, "")
	if err != nil {
		logger.LogError(err)
		tx.Rollback()
		return err
	}

	if len(commits) == 0 {
		logger.LogInfo("No new commits found")
		return nil
	}

	// Save new commits
	if err := s.SaveCommits(ctx, commits); err != nil {
		logger.LogError(err)
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.LogError(err)
		return err
	}

	return nil
}
