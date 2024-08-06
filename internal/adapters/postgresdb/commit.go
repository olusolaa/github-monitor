package postgresdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/pagination"
)

type commitRepository struct {
	db *sqlx.DB
}

type CommitRepository interface {
	Save(ctx context.Context, commits []domain.Commit) error
	GetLatestCommitByRepositoryID(ctx context.Context, repoID int64) (*domain.Commit, error)
	GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, int, error)
	DeleteCommitsByRepositoryID(ctx context.Context, repoID int64) error
	GetTopCommitAuthors(ctx context.Context, owner, name string, limit int) ([]domain.CommitAuthor, error)
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
}

func NewCommitRepository(db *sqlx.DB) CommitRepository {
	return &commitRepository{db: db}
}

// Save inserts new commits into the database. Commits with existing hashes are ignored.
func (c commitRepository) Save(ctx context.Context, commits []domain.Commit) error {
	query := `
        INSERT INTO commits (repository_id, hash, message, author_name, author_email, commit_date, url)
        VALUES (:repository_id, :hash, :message, :author_name, :author_email, :commit_date, :url)
        ON CONFLICT (hash) DO NOTHING;
    `
	_, err := c.db.NamedExecContext(ctx, query, commits)
	if err != nil {
		return fmt.Errorf("database save error: %w", err)
	}
	return nil
}

// GetLatestCommitByRepositoryID retrieves the most recent commit for a specified repository.
func (c commitRepository) GetLatestCommitByRepositoryID(ctx context.Context, repoID int64) (*domain.Commit, error) {
	query := `
        SELECT id, repository_id, hash, message, author_name, author_email, commit_date, url
        FROM commits
        WHERE repository_id = $1
        ORDER BY commit_date DESC
        LIMIT 1;
    `
	var commit domain.Commit
	if err := c.db.GetContext(ctx, &commit, query, repoID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest commit: %w", err)
	}
	return &commit, nil
}

func (c commitRepository) GetCommitsByRepositoryName(ctx context.Context, owner, name string, page, pageSize int) ([]domain.Commit, int, error) {
	query := `
        SELECT commits.id, commits.repository_id, commits.hash, commits.message, commits.author_name, 
               commits.author_email, commits.commit_date, commits.url
        FROM commits
        JOIN repositories ON commits.repository_id = repositories.id
        WHERE repositories.name = $1 AND repositories.owner = $2
        ORDER BY commits.commit_date DESC
    `
	paginatedQuery := pagination.ApplyToQuery(query, page, pageSize)

	var commits []domain.Commit
	if err := c.db.SelectContext(ctx, &commits, paginatedQuery, name, owner); err != nil {
		return nil, 0, fmt.Errorf("failed to get commits by repository name: %w", err)
	}

	// Count total items for pagination
	var totalItems int
	countQuery := `SELECT COUNT(*) FROM commits
                   JOIN repositories ON commits.repository_id = repositories.id
                   WHERE repositories.name = $1 AND repositories.owner = $2`
	if err := c.db.GetContext(ctx, &totalItems, countQuery, name, owner); err != nil {
		return nil, 0, fmt.Errorf("failed to count total commits: %w", err)
	}

	return commits, totalItems, nil
}

// DeleteCommitsByRepositoryID deletes all commits for a specified repository.
func (c commitRepository) DeleteCommitsByRepositoryID(ctx context.Context, repoID int64) error {
	query := `
        DELETE FROM commits
        WHERE repository_id = $1;
    `
	if _, err := c.db.ExecContext(ctx, query, repoID); err != nil {
		return fmt.Errorf("failed to delete commits: %w", err)
	}
	return nil
}

// GetTopCommitAuthors retrieves the top N authors by commit count for a specified repository.
func (c commitRepository) GetTopCommitAuthors(ctx context.Context, owner, name string, limit int) ([]domain.CommitAuthor, error) {
	query := `
        SELECT c.author_name, c.author_email, COUNT(*) AS commit_count
        FROM commits c
        INNER JOIN repositories r ON c.repository_id = r.id
        WHERE r.name = $1 AND r.owner = $2
        GROUP BY c.author_name, c.author_email
        ORDER BY commit_count DESC
        LIMIT $3;
    `
	var authors []domain.CommitAuthor
	if err := c.db.SelectContext(ctx, &authors, query, name, owner, limit); err != nil {
		return nil, fmt.Errorf("failed to get top commit authors: %w", err)
	}
	return authors, nil
}

// BeginTx starts a new database transaction.
func (c commitRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}
