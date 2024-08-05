package services

import (
	"context"
	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/adapters/queue"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
	"strconv"
)

type RepositoryService interface {
	GetRepository(ctx context.Context, name, owner string) (*domain.Repository, error)
	GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error)
	UpsertRepository(ctx context.Context, repository *domain.Repository) error
	InitializeRepository(publisher queue.MessagePublisher, owner, repo string) error
}

type repositoryService struct {
	ghService GitHubService
	repoRepo  *postgresdb.RepositoryRepository
}

func NewRepositoryService(ghService GitHubService, repoRepo *postgresdb.RepositoryRepository) RepositoryService {
	return &repositoryService{ghService: ghService, repoRepo: repoRepo}
}

func (s *repositoryService) InitializeRepository(publisher queue.MessagePublisher, owner, repo string) error {
	ctx := context.Background()

	repository, err := s.ghService.FetchRepository(ctx, repo, owner)
	if err != nil {
		logger.LogError(err)
		return err
	}

	err = s.UpsertRepository(ctx, repository)
	if err != nil {
		logger.LogError(err)
		return err
	}

	err = publisher.PublishMessage("fetch_commits", strconv.Itoa(int(repository.ID)))
	if err != nil {
		logger.LogError(err)
		return err
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
