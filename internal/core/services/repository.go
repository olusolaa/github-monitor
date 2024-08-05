package services

import (
	"context"
	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type RepositoryService interface {
	GetRepository(ctx context.Context, name, owner string) (*domain.Repository, error)
	GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error)
	UpsertRepository(ctx context.Context, repository *domain.Repository) error
	AddRepository(ctx context.Context, owner, repo string) error
	FetchRepository(ctx context.Context, owner, repo string, commitChan chan int64) error
}

type RepoRequest struct {
	Owner string
	Name  string
	retry int
	ctx   context.Context
}

type repositoryService struct {
	ghService GitHubService
	repoRepo  *postgresdb.RepositoryRepository
	repoChan  chan RepoRequest
}

func NewRepositoryService(ghService GitHubService, repoRepo *postgresdb.RepositoryRepository, commitChan chan int64) RepositoryService {
	s := &repositoryService{
		ghService: ghService,
		repoRepo:  repoRepo,
		repoChan:  make(chan RepoRequest, 10), // Buffered channel for concurrent requests
	}
	go s.RepositoryManager(commitChan) // Start the manager goroutine with the commitChan
	return s
}

func (s *repositoryService) AddRepository(ctx context.Context, owner, repo string) error {
	repoRequest := RepoRequest{
		Owner: owner,
		Name:  repo,
		retry: 0,
		ctx:   ctx,
	}
	s.repoChan <- repoRequest
	return nil
}

func (s *repositoryService) RepositoryManager(commitChan chan int64) {
	for {
		select {
		case repoRequest := <-s.repoChan:
			err := s.FetchRepository(repoRequest.ctx, repoRequest.Owner, repoRequest.Name, commitChan)
			if err != nil {
				repoRequest.retry++
				if repoRequest.retry < 3 {
					s.repoChan <- repoRequest
				} else {
					logger.LogError(err)
				}
			}
		}
	}
}

func (s *repositoryService) FetchRepository(ctx context.Context, owner, repo string, commitChan chan int64) error {
	repository, err := s.ghService.FetchRepository(ctx, owner, repo)
	if err != nil {
		logger.LogError(err)
		return err
	}

	err = s.UpsertRepository(ctx, repository)
	if err != nil {
		logger.LogError(err)
		return err
	}

	if commitChan != nil {
		commitChan <- repository.ID
	}
	logger.LogInfo("Initialized repository and published event for fetching commits")
	return nil
}

// GetRepositoryInfo fetches repository information either from the database or GitHub API.
func (s *repositoryService) GetRepository(ctx context.Context, repoName, owner string) (*domain.Repository, error) {

	repository, err := s.repoRepo.FindByNameAndOwner(ctx, repoName, owner)
	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	return repository, nil
}

func (s *repositoryService) UpsertRepository(ctx context.Context, repository *domain.Repository) error {
	if err := s.repoRepo.Upsert(ctx, repository); err != nil {
		logger.LogError(err)
		return err
	}
	return nil
}

func (s *repositoryService) GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error) {
	owner, repoName, err := s.repoRepo.GetOwnerAndRepoName(ctx, repoID)
	if err != nil {
		return "", "", err
	}
	return owner, repoName, nil
}
