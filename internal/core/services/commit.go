package services

import (
	"context"
	"fmt"
	"time"

	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/errors"
	"github.com/olusolaa/github-monitor/pkg/logger"
	"github.com/olusolaa/github-monitor/pkg/pagination"
)

type CommitService interface {
	SaveCommits(ctx context.Context, commits []domain.Commit) error
	GetLatestCommit(ctx context.Context, repoID int64) (*domain.Commit, error)
	GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, *pagination.Pagination, error)
	ResetCollection(ctx context.Context, owner, name string, startTime time.Time) error
	GetTopCommitAuthors(ctx context.Context, owner, name string, limit int) ([]domain.CommitAuthor, error)
	CommitManager(monitoringChan chan int64, startDate, endDate string)
	ProcessCommits(repoID int64, monitoringChan chan int64, startDate, endDate string)
}

type commitService struct {
	gitHubService     GitHubService
	repositoryService RepositoryService
	commitRepo        postgresdb.CommitRepository
	commitChan        chan int64
}

func NewCommitService(gitHubService GitHubService, repositoryService RepositoryService, commitRepo postgresdb.CommitRepository, commitChan chan int64) CommitService {
	return &commitService{
		gitHubService:     gitHubService,
		repositoryService: repositoryService,
		commitRepo:        commitRepo,
		commitChan:        commitChan,
	}
}

func (cs *commitService) CommitManager(monitoringChan chan int64, startDate, endDate string) {
	for repoID := range cs.commitChan {
		go cs.ProcessCommits(repoID, monitoringChan, startDate, endDate)
	}
}

func (cs *commitService) ProcessCommits(repoID int64, monitoringChan chan int64, startDate, endDate string) {
	ctx := context.Background()

	owner, name, err := cs.repositoryService.GetOwnerAndRepoName(ctx, repoID)
	if err != nil {
		logger.LogError(errors.New("GET_OWNER_REPO_NAME_ERROR", "error getting owner and repo name", err, errors.Critical))
		return
	}

	var encounteredError error

	domainCommitsChan := make(chan []domain.Commit)
	errChan := make(chan error)

	defer func() {
		close(domainCommitsChan)
		close(errChan)

		if encounteredError != nil {
			logger.LogError(errors.New("FETCH_COMMITS_ERROR", "error fetching commits", encounteredError, errors.Critical))
		} else {
			logger.LogInfo(fmt.Sprintf("Commits fetched successfully for %s/%s", owner, name))
			monitoringChan <- repoID
		}
	}()

	go cs.gitHubService.FetchCommits(ctx, owner, name, startDate, endDate, repoID, domainCommitsChan, errChan)

	for {
		select {
		case domainCommits, ok := <-domainCommitsChan:
			if !ok {
				encounteredError = errors.New("DOMAIN_COMMITS_CHANNEL_CLOSED", "domain commits channel closed unexpectedly", nil, errors.Critical)
				return
			}
			if err := cs.SaveCommits(ctx, domainCommits); err != nil {
				encounteredError = err
				return
			}
		case err, ok := <-errChan:
			if !ok {
				encounteredError = errors.New("ERR_CHANNEL_CLOSED", "error channel closed unexpectedly", nil, errors.Critical)
			} else if err != nil {
				encounteredError = err
			}
			return
		case <-ctx.Done():
			encounteredError = errors.New("CONTEXT_DONE", "context canceled or timed out", ctx.Err(), errors.Critical)
			return
		}
	}
}

// SaveCommits saves the provided commits into the repository
func (s *commitService) SaveCommits(ctx context.Context, commits []domain.Commit) error {
	if err := s.commitRepo.Save(ctx, commits); err != nil {
		logger.LogError(errors.New("SAVE_COMMITS_ERROR", "error saving commits", err, errors.Critical))
		return err
	}
	logger.LogInfo(fmt.Sprintf("Saved %d commits successfully", len(commits)))
	return nil
}

// GetLatestCommit retrieves the most recent commit for a given repository
func (s *commitService) GetLatestCommit(ctx context.Context, repoID int64) (*domain.Commit, error) {
	latestCommit, err := s.commitRepo.GetLatestCommitByRepositoryID(ctx, repoID)
	if err != nil {
		logger.LogError(errors.New("GET_LATEST_COMMIT_ERROR", "error retrieving the latest commit", err, errors.Critical))
		return nil, err
	}
	return latestCommit, nil
}

func (s *commitService) GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, *pagination.Pagination, error) {
	commits, totalItems, err := s.commitRepo.GetCommitsByRepositoryName(ctx, owner, name, page, pageSize)
	if err != nil {
		logger.LogError(errors.New("GET_COMMITS_ERROR", "error retrieving commits", err, errors.Critical))
		return nil, nil, err
	}

	pg := pagination.NewPagination(page, pageSize, totalItems)
	logger.LogInfo(fmt.Sprintf("Fetched %d commits for %s/%s", len(commits), owner, name))
	return commits, pg, nil
}

func (s *commitService) GetTopCommitAuthors(ctx context.Context, owner, name string, limit int) ([]domain.CommitAuthor, error) {
	authors, err := s.commitRepo.GetTopCommitAuthors(ctx, owner, name, limit)
	if err != nil {
		logger.LogError(errors.New("GET_TOP_AUTHORS_ERROR", "error retrieving top commit authors", err, errors.Critical))
		return nil, err
	}
	logger.LogInfo(fmt.Sprintf("Fetched top commit authors for repo ID: %s", name))
	return authors, nil
}

func (s *commitService) ResetCollection(ctx context.Context, owner, name string, startTime time.Time) error {
	tx, err := s.commitRepo.BeginTx(ctx)
	if err != nil {
		logger.LogError(errors.New("BEGIN_TRANSACTION_ERROR", "error beginning transaction", err, errors.Critical))
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.LogError(errors.New("PANIC", "panic occurred during ResetCollection", fmt.Errorf("%v", r), errors.Critical))
		}
	}()

	rep, err := s.repositoryService.GetRepository(ctx, name, owner)
	if err != nil {
		logger.LogError(errors.New("DELETE_COMMITS_ERROR", "error deleting commits", err, errors.Critical))
		tx.Rollback()
		return err
	}

	startTimeStr := startTime.Format(time.RFC3339)
	err = s.commitRepo.DeleteCommitsByRepositoryID(ctx, rep.ID)
	if err != nil {
		logger.LogError(errors.New("GET_OWNER_REPO_NAME_ERROR", "error getting owner and repo name", err, errors.Critical))
		tx.Rollback()
		return err
	} else {
		if err := tx.Commit(); err != nil {
			logger.LogError(errors.New("COMMIT_TRANSACTION_ERROR", "error committing transaction", err, errors.Critical))
			return err
		}
		logger.LogInfo(fmt.Sprintf("Collection reset successfully for repository name: %s", name))
	}

	domainCommitsChan := make(chan []domain.Commit)
	errChan := make(chan error)

	defer func() {
		close(domainCommitsChan)
		close(errChan)
	}()

	go s.gitHubService.FetchCommits(ctx, rep.Owner, rep.Name, startTimeStr, "", rep.ID, domainCommitsChan, errChan)

	for {
		select {
		case domainCommits, ok := <-domainCommitsChan:
			if !ok {
				return errors.New("DOMAIN_COMMITS_CHANNEL_CLOSED", "domain commits channel closed unexpectedly", nil, errors.Critical)
			}
			if err := s.SaveCommits(ctx, domainCommits); err != nil {
				logger.LogError(errors.New("SAVE_COMMITS_ERROR", "error saving commits", err, errors.Critical))
				tx.Rollback()
				return err
			}
		case err, ok := <-errChan:
			if !ok {
				return errors.New("ERR_CHANNEL_CLOSED", "error channel closed unexpectedly", nil, errors.Critical)
			}
			if err != nil {
				logger.LogError(errors.New("FETCH_COMMITS_ERROR", "error fetching commits", err, errors.Critical))
				tx.Rollback()
			}
			return err
		case <-ctx.Done():
			logger.LogError(errors.New("CONTEXT_DONE", "context canceled or timed out", ctx.Err(), errors.Critical))
			tx.Rollback()
			return ctx.Err()
		}
	}
}
