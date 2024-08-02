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
	//GetRepoFromDB(ctx context.Context, repoID int64) (*domain.Repository, error)
}

type repositoryService struct {
	ghService GitHubService
	repoRepo  *postgresdb.RepositoryRepository
}

func NewRepositoryService(ghService GitHubService, repoRepo *postgresdb.RepositoryRepository) RepositoryService {
	return &repositoryService{ghService: ghService, repoRepo: repoRepo}
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
