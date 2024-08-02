package postgresdb

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/olusolaa/github-monitor/internal/core/domain"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type CommitRepository struct {
	db *sqlx.DB
}

func NewCommitRepository(db *sqlx.DB) *CommitRepository {
	return &CommitRepository{db: db}
}

// Save inserts new commits into the database. Commits with existing hashes are ignored.
func (c *CommitRepository) Save(ctx context.Context, commits []domain.Commit) error {
	query := `
        INSERT INTO commits (repository_id, hash, message, author_name, author_email, commit_date, url)
        VALUES (:repository_id, :hash, :message, :author_name, :author_email, :commit_date, :url)
        ON CONFLICT (hash) DO NOTHING;
    `
	if _, err := c.db.NamedExecContext(ctx, query, commits); err != nil {
		logger.LogError(err)
		return err
	}
	return nil
}

// GetLatestCommitByRepositoryID retrieves the most recent commit for a specified repository.
func (c *CommitRepository) GetLatestCommitByRepositoryID(ctx context.Context, repoID int64) (*domain.Commit, error) {
	query := `
        SELECT id, repository_id, hash, message, author_name, author_email, commit_date, url
        FROM commits
        WHERE repository_id = $1
        ORDER BY commit_date DESC
        LIMIT 1;
    `
	var commit domain.Commit
	if err := c.db.GetContext(ctx, &commit, query, repoID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.LogError(err)
		return nil, err
	}
	return &commit, nil
}

// GetCommitsByRepositoryID retrieves all commits for a specified repository, ordered by commit date.
func (c *CommitRepository) GetCommitsByRepositoryID(ctx context.Context, repoID int64) ([]domain.Commit, error) {
	query := `
        SELECT id, repository_id, hash, message, author_name, author_email, commit_date, url
        FROM commits
        WHERE repository_id = $1
        ORDER BY commit_date DESC;
    `
	var commits []domain.Commit
	if err := c.db.SelectContext(ctx, &commits, query, repoID); err != nil {
		logger.LogError(err)
		return nil, err
	}
	return commits, nil
}

// DeleteCommitsByRepositoryID deletes all commits for a specified repository.
func (c *CommitRepository) DeleteCommitsByRepositoryID(ctx context.Context, repoID int64) error {
	query := `
        DELETE FROM commits
        WHERE repository_id = $1;
    `
	if _, err := c.db.ExecContext(ctx, query, repoID); err != nil {
		logger.LogError(err)
		return err
	}
	return nil
}

// GetTopCommitAuthors retrieves the top N authors by commit count for a specified repository.
func (c *CommitRepository) GetTopCommitAuthors(ctx context.Context, repoID int64, limit int) ([]domain.CommitAuthor, error) {
	query := `
        SELECT author_name, author_email, COUNT(*) AS commit_count
        FROM commits
        WHERE repository_id = $1
        GROUP BY author_name, author_email
        ORDER BY commit_count DESC
        LIMIT $2;
    `
	var authors []domain.CommitAuthor
	if err := c.db.SelectContext(ctx, &authors, query, repoID, limit); err != nil {
		logger.LogError(err)
		return nil, err
	}
	return authors, nil
}

// BeginTx starts a new database transaction.
func (c *CommitRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	tx, err := c.db.BeginTxx(ctx, nil)
	if err != nil {
		logger.LogError(err)
		return nil, err
	}
	return tx, nil
}
