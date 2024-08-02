package services

import (
	"context"
	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type RepositoryService interface {
	FetchRepositoryInfo(ctx context.Context, name, owner string) (*domain.Repository, error)
	GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error)
}

type repositoryService struct {
	ghService GitHubService
	repoRepo  *postgresdb.RepositoryRepository
}

func NewRepositoryService(ghService GitHubService, repoRepo *postgresdb.RepositoryRepository) RepositoryService {
	return &repositoryService{ghService: ghService, repoRepo: repoRepo}
}

func (s *repositoryService) FetchRepositoryInfo(ctx context.Context, repoName, owner string) (*domain.Repository, error) {

	repository, err := s.repoRepo.FindByNameAndOwner(ctx, repoName, owner)
	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	// This seems incorrect. The repository should be fetched from GitHub if it doesn't exist in the database
	if repository != nil {
		return repository, nil
	}

	// Fetch latest repository data from GitHub
	repository, err = s.ghService.FetchRepository(ctx, owner, repoName)
	if err != nil {
		logger.LogError(err)
		return nil, err
	}

	repository.Owner = owner

	if err := s.repoRepo.Upsert(ctx, repository); err != nil {
		logger.LogError(err)
		return nil, err
	}

	return repository, nil
}

func (s *repositoryService) GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error) {
	owner, repoName, err := s.repoRepo.GetOwnerAndRepoName(ctx, repoID)
	if err != nil {
		return "", "", err
	}
	return owner, repoName, nil
}
