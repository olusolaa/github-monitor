package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
	"github.com/olusolaa/github-monitor/pkg/pagination"
)

type CommitService interface {
	SaveCommits(ctx context.Context, commits []domain.Commit) error
	GetLatestCommit(ctx context.Context, repoID int64) (*domain.Commit, error)
	GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, *pagination.Pagination, error)
	ResetCollection(ctx context.Context, repoID int64, startTime time.Time) error
	GetTopCommitAuthors(ctx context.Context, repoID int64, limit int) ([]domain.CommitAuthor, error)
	CommitManager(monitoringChan chan int64, startDate, endDate string)
}

type commitService struct {
	gitHubService     GitHubService
	repositoryService RepositoryService
	commitRepo        *postgresdb.CommitRepository
	commitChan        chan int64
}

func NewCommitService(gitHubService GitHubService, repositoryService RepositoryService, commitRepo *postgresdb.CommitRepository, commitChan chan int64) CommitService {
	return &commitService{
		gitHubService:     gitHubService,
		repositoryService: repositoryService,
		commitRepo:        commitRepo,
		commitChan:        commitChan,
	}
}

func (cs *commitService) CommitManager(monitoringChan chan int64, startDate, endDate string) {
	for repoID := range cs.commitChan {
		go cs.processCommits(repoID, monitoringChan, startDate, endDate)
	}
}

func (cs *commitService) processCommits(repoID int64, monitoringChan chan int64, startDate, endDate string) {
	ctx := context.Background()

	owner, name, err := cs.repositoryService.GetOwnerAndRepoName(ctx, repoID)
	if err != nil {
		log.Printf("Error getting owner and repo name: %v", err)
		return
	}

	domainCommitsChan := make(chan []domain.Commit)
	errChan := make(chan error)

	go cs.gitHubService.FetchCommits(ctx, owner, name, startDate, endDate, repoID, domainCommitsChan, errChan)

	var fetchError error

	for {
		select {
		case domainCommits, ok := <-domainCommitsChan:
			if !ok {
				domainCommitsChan = nil
			} else {
				if err := cs.SaveCommits(ctx, domainCommits); err != nil {
					log.Printf("Error saving commits: %v", err)
					if fetchError == nil {
						fetchError = err
					}
				}
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if err != nil && fetchError == nil {
				fetchError = fmt.Errorf("error fetching commits: %w", err)
			}
		case <-ctx.Done():
			if fetchError == nil {
				fetchError = ctx.Err()
			}
		}

		if domainCommitsChan == nil && errChan == nil {
			break
		}
	}

	if fetchError == nil {
		monitoringChan <- repoID
	} else {
		log.Printf("Error during commit processing: %v", fetchError)
	}
	close(domainCommitsChan)
	close(errChan)
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
