package services

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/pkg/pagination"
	"time"

	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type CommitService interface {
	SaveCommits(ctx context.Context, commits []domain.Commit) error
	GetLatestCommit(ctx context.Context, repoID int64) (*domain.Commit, error)
	GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, *pagination.Pagination, error)
	ResetCollection(ctx context.Context, repoID int64, startTime time.Time) error
	GetTopCommitAuthors(ctx context.Context, repoID int64, limit int) ([]domain.CommitAuthor, error)
}

type commitService struct {
	gitHubService     GitHubService
	repositoryService RepositoryService
	commitRepo        *postgresdb.CommitRepository
}

func NewCommitService(gitHubService GitHubService, repositoryService RepositoryService, commitRepo *postgresdb.CommitRepository) CommitService {
	return &commitService{gitHubService: gitHubService, repositoryService: repositoryService, commitRepo: commitRepo}
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

func (s *commitService) GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, *pagination.Pagination, error) {
	commits, totalItems, err := s.commitRepo.GetCommitsByRepositoryName(ctx, owner, name, page, pageSize)
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
	owner, name, err := s.repositoryService.GetOwnerAndRepoName(ctx, repoID)
	if err != nil {
		logger.LogError(err)
		tx.Rollback()
		return err
	}

	domainCommitsChan := make(chan []domain.Commit)
	errChan := make(chan error)

	go s.gitHubService.FetchCommits(ctx, owner, name, startTimeStr, "", repoID, domainCommitsChan, errChan)

	for {
		select {
		case domainCommits, ok := <-domainCommitsChan:
			if !ok {
				domainCommitsChan = nil
			} else {
				if err := s.SaveCommits(ctx, domainCommits); err != nil {
					logger.LogError(err)
					tx.Rollback()
					return err
				}
			}
		case err := <-errChan:
			logger.LogError(err)
			tx.Rollback()
			return fmt.Errorf("error fetching commits: %w", err)
		case <-ctx.Done():
			tx.Rollback()
			return ctx.Err()
		}

		if domainCommitsChan == nil && errChan == nil {
			break
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.LogError(err)
		return err
	}

	return nil
}
