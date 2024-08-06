package test

import (
	"context"
	"github.com/jmoiron/sqlx"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/pkg/pagination"
)

// Mock types
type MockGitHubService struct{ mock.Mock }

type MockRepositoryService struct{ mock.Mock }
type MockCommitRepository struct{ mock.Mock }

func (m *MockGitHubService) FetchCommits(ctx context.Context, owner, name, startDate, endDate string, repoID int64, domainCommitsChan chan<- []domain.Commit, errChan chan<- error) {
	m.Called(ctx, owner, name, startDate, endDate, repoID, domainCommitsChan, errChan)
}

func (m *MockGitHubService) FetchRepository(ctx context.Context, owner string, repoName string) (*domain.Repository, error) {
	args := m.Called(ctx, owner, repoName)
	return args.Get(0).(*domain.Repository), args.Error(1)
}

func (m *MockRepositoryService) GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error) {
	args := m.Called(ctx, repoID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockRepositoryService) GetRepository(ctx context.Context, name string, owner string) (*domain.Repository, error) {
	args := m.Called(ctx, name, owner)
	return args.Get(0).(*domain.Repository), args.Error(1)
}

func (m *MockRepositoryService) UpsertRepository(ctx context.Context, repository *domain.Repository) error {
	args := m.Called(ctx, repository)
	return args.Error(0)
}

func (m *MockRepositoryService) AddRepository(owner string, repo string) error {
	args := m.Called(owner, repo)
	return args.Error(0)
}

func (m *MockRepositoryService) FetchRepository(ctx context.Context, owner string, repo string, commitChan chan int64) error {
	args := m.Called(ctx, owner, repo, commitChan)
	return args.Error(0)
}

func (m *MockCommitRepository) Save(ctx context.Context, commits []domain.Commit) error {
	args := m.Called(ctx, commits)
	return args.Error(0)
}

func (m *MockCommitRepository) GetLatestCommitByRepositoryID(ctx context.Context, repoID int64) (*domain.Commit, error) {
	args := m.Called(ctx, repoID)
	return args.Get(0).(*domain.Commit), args.Error(1)
}

func (m *MockCommitRepository) GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, int, error) {
	args := m.Called(ctx, owner, name, page, pageSize)
	return args.Get(0).([]domain.Commit), args.Int(1), args.Error(2)
}

func (m *MockCommitRepository) GetTopCommitAuthors(ctx context.Context, repoID int64, limit int) ([]domain.CommitAuthor, error) {
	args := m.Called(ctx, repoID, limit)
	return args.Get(0).([]domain.CommitAuthor), args.Error(1)
}

func (m *MockCommitRepository) DeleteCommitsByRepositoryID(ctx context.Context, repoID int64) error {
	args := m.Called(ctx, repoID)
	return args.Error(0)
}

func (m *MockCommitRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(*sqlx.Tx), args.Error(1)
}

// Test cases
func TestCommitService_SaveCommits(t *testing.T) {
	mockGitHubService := new(MockGitHubService)
	mockRepoService := new(MockRepositoryService)
	mockCommitRepo := new(MockCommitRepository)
	mockCommitChan := make(chan int64)

	service := services.NewCommitService(mockGitHubService, mockRepoService, mockCommitRepo, mockCommitChan)

	commits := []domain.Commit{{RepositoryID: 1, Hash: "hash1", Message: "Commit message", CommitDate: time.Now()}}

	mockCommitRepo.On("Save", mock.Anything, commits).Return(nil)

	err := service.SaveCommits(context.Background(), commits)

	assert.NoError(t, err)
	mockCommitRepo.AssertExpectations(t)
}

func TestCommitService_GetLatestCommit(t *testing.T) {
	mockGitHubService := new(MockGitHubService)
	mockRepoService := new(MockRepositoryService)
	mockCommitRepo := new(MockCommitRepository)
	mockCommitChan := make(chan int64)

	service := services.NewCommitService(mockGitHubService, mockRepoService, mockCommitRepo, mockCommitChan)

	expectedCommit := &domain.Commit{RepositoryID: 1, Hash: "hash1", Message: "Commit message", CommitDate: time.Now()}

	mockCommitRepo.On("GetLatestCommitByRepositoryID", mock.Anything, int64(1)).Return(expectedCommit, nil)

	commit, err := service.GetLatestCommit(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, expectedCommit, commit)
	mockCommitRepo.AssertExpectations(t)
}

func TestCommitService_GetCommitsByRepositoryName(t *testing.T) {
	mockGitHubService := new(MockGitHubService)
	mockRepoService := new(MockRepositoryService)
	mockCommitRepo := new(MockCommitRepository)
	mockCommitChan := make(chan int64)

	service := services.NewCommitService(mockGitHubService, mockRepoService, mockCommitRepo, mockCommitChan)

	expectedCommits := []domain.Commit{{RepositoryID: 1, Hash: "hash1", Message: "Commit message", CommitDate: time.Now()}}
	totalItems := 1

	mockCommitRepo.On("GetCommitsByRepositoryName", mock.Anything, "owner", "name", 1, 10).Return(expectedCommits, totalItems, nil)

	commits, pg, err := service.GetCommitsByRepositoryName(context.Background(), "owner", "name", 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, expectedCommits, commits)
	assert.Equal(t, pagination.NewPagination(1, 10, totalItems), pg)
	mockCommitRepo.AssertExpectations(t)
}

func TestCommitService_GetTopCommitAuthors(t *testing.T) {
	mockGitHubService := new(MockGitHubService)
	mockRepoService := new(MockRepositoryService)
	mockCommitRepo := new(MockCommitRepository)
	mockCommitChan := make(chan int64)

	service := services.NewCommitService(mockGitHubService, mockRepoService, mockCommitRepo, mockCommitChan)

	expectedAuthors := []domain.CommitAuthor{
		{AuthorName: "John Doe", AuthorEmail: "john@example.com", CommitCount: 5},
		{AuthorName: "Jane Doe", AuthorEmail: "jane@example.com", CommitCount: 3},
	}

	mockCommitRepo.On("GetTopCommitAuthors", mock.Anything, int64(1), 2).Return(expectedAuthors, nil)

	authors, err := service.GetTopCommitAuthors(context.Background(), 1, 2)

	assert.NoError(t, err)
	assert.Equal(t, expectedAuthors, authors)
	mockCommitRepo.AssertExpectations(t)
}

func TestCommitService_ProcessCommits(t *testing.T) {
	mockGitHubService := new(MockGitHubService)
	mockRepoService := new(MockRepositoryService)
	mockCommitRepo := new(MockCommitRepository)

	repoID := int64(1)
	owner := "testOwner"
	name := "testRepo"
	startDate := "2022-01-01"
	endDate := "2022-01-31"
	commits := []domain.Commit{
		{Hash: "abc123", Message: "Initial commit"},
	}

	// Setup expectations
	mockRepoService.On("GetOwnerAndRepoName", mock.Anything, repoID).Return(owner, name, nil)
	mockGitHubService.On("FetchCommits", mock.Anything, owner, name, startDate, endDate, repoID, mock.AnythingOfType("chan<- []domain.Commit"), mock.AnythingOfType("chan<- error")).Run(func(args mock.Arguments) {
		commitsChan := args.Get(6).(chan<- []domain.Commit)
		errChan := args.Get(7).(chan<- error)

		go func() {
			commitsChan <- commits
			errChan <- nil
		}()
	}).Return(nil)

	mockCommitRepo.On("Save", mock.Anything, commits).Return(nil)

	monitoringChan := make(chan int64, 1)

	cs := services.NewCommitService(mockGitHubService, mockRepoService, mockCommitRepo, monitoringChan)

	cs.ProcessCommits(repoID, monitoringChan, startDate, endDate)

	mockRepoService.AssertExpectations(t)
	mockGitHubService.AssertExpectations(t)
	mockCommitRepo.AssertExpectations(t)

	select {
	case id := <-monitoringChan:
		assert.Equal(t, repoID, id)
	case <-time.After(time.Second):
		t.Fatal("expected repoID in monitoringChan")
	}
}
