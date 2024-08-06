package postgresdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"time"
)

type repositoryRepository struct {
	db *sqlx.DB
}

type RepositoryRepository interface {
	Upsert(ctx context.Context, repository *domain.Repository) error
	FindByNameAndOwner(ctx context.Context, name, owner string) (*domain.Repository, error)
	GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error)
	Update(ctx context.Context, repo *domain.Repository) error
}

func NewRepositoryRepository(db *sqlx.DB) RepositoryRepository {
	return &repositoryRepository{db: db}
}

// Upsert inserts or updates a repository record in the database.
func (r *repositoryRepository) Upsert(ctx context.Context, repository *domain.Repository) error {
	query := `
        INSERT INTO repositories (name, owner, description, url, language, forks_count, stargazers_count, open_issues_count, watchers_count, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (name, owner) DO UPDATE SET
            description = EXCLUDED.description,
            url = EXCLUDED.url,
            language = EXCLUDED.language,
            forks_count = EXCLUDED.forks_count,
            stargazers_count = EXCLUDED.stargazers_count,
            open_issues_count = EXCLUDED.open_issues_count,
            watchers_count = EXCLUDED.watchers_count,
            updated_at = EXCLUDED.updated_at
        RETURNING id;
    `
	err := r.db.QueryRowContext(ctx, query,
		repository.Name,
		repository.Owner,
		repository.Description,
		repository.URL,
		repository.Language,
		repository.ForksCount,
		repository.StargazersCount,
		repository.OpenIssuesCount,
		repository.WatchersCount,
		repository.CreatedAt,
		repository.UpdatedAt,
	).Scan(&repository.ID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to upsert repository: %w", err)
		}
		return err
	}
	return nil
}

// FindByNameAndOwner retrieves a repository by its name and owner.
func (r repositoryRepository) FindByNameAndOwner(ctx context.Context, name, owner string) (*domain.Repository, error) {
	query := `SELECT id, name, owner, description, url, language, forks_count, stargazers_count, open_issues_count, watchers_count, created_at, updated_at FROM repositories WHERE name = $1 AND owner = $2`
	var repository domain.Repository
	err := r.db.GetContext(ctx, &repository, query, name, owner)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No rows found, return nil without logging an error
		}
		return nil, fmt.Errorf("failed to find repository by name and owner: %w", err)
	}
	return &repository, nil
}

// GetOwnerAndRepoName retrieves the owner and repository name by repository ID.
func (r repositoryRepository) GetOwnerAndRepoName(ctx context.Context, repoID int64) (string, string, error) {
	query := `SELECT owner, name FROM repositories WHERE id = $1`
	var owner, name string
	err := r.db.QueryRowContext(ctx, query, repoID).Scan(&owner, &name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil // No rows found, return empty strings without error
		}
		return "", "", fmt.Errorf("failed to get owner and repository name: %w", err)
	}
	return owner, name, nil
}

// Update updates the repository record in the database.
func (r repositoryRepository) Update(ctx context.Context, repo *domain.Repository) error {
	query := `UPDATE repositories SET
        forks_count = $1,
        stargazers_count = $2,
        open_issues_count = $3,
        watchers_count = $4,
        updated_at = $5
        WHERE id = $6`

	_, err := r.db.ExecContext(ctx, query, repo.ForksCount, repo.StargazersCount, repo.OpenIssuesCount, repo.WatchersCount, time.Now(), repo.ID)
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}
	return nil
}
