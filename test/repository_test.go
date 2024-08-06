package test

import (
	"context"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockRepositoryRepository struct {
	mock.Mock
}

func (m *MockRepositoryRepository) FindByNameAndOwner(ctx context.Context, name, owner string) (*domain.Repository, error) {
	args := m.Called(ctx, name, owner)
	return args.Get(0).(*domain.Repository), args.Error(1)
}

func (m *MockRepositoryRepository) Upsert(ctx context.Context, repository *domain.Repository) error {
	args := m.Called(ctx, repository)
	return args.Error(0)
}

func (m *MockRepositoryRepository) GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error) {
	args := m.Called(ctx, repoID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockRepositoryRepository) Update(ctx context.Context, repo *domain.Repository) error {
	args := m.Called(ctx, repo)
	return args.Error(0)
}

func (m *MockRepositoryService) RepositoryManager(commitChan chan int64) {
	m.Called(commitChan)
}

func TestRepositoryManager(t *testing.T) {
	mockGHService := new(MockGitHubService)
	mockRepoRepo := new(MockRepositoryRepository)
	commitChan := make(chan int64, 1)
	repoChan := make(chan services.RepoRequest, 1)
	service := services.NewRepositoryService(mockGHService, mockRepoRepo, repoChan, commitChan)

	repo := &domain.Repository{ID: 1, Name: "testRepo", Owner: "testOwner"}
	mockGHService.On("FetchRepository", mock.Anything, "testOwner", "testRepo").Return(repo, nil).Once()
	mockRepoRepo.On("Upsert", mock.Anything, repo).Return(nil)

	go service.RepositoryManager(commitChan)

	repoChan <- services.RepoRequest{Owner: "testOwner", Name: "testRepo"}

	select {
	case id := <-commitChan:
		assert.Equal(t, int64(1), id)
	case <-time.After(time.Second): // Adjust timeout as necessary
		t.Fatal("expected repoID in commitChan")
	}

	mockGHService.AssertExpectations(t)
	mockRepoRepo.AssertExpectations(t)
}
