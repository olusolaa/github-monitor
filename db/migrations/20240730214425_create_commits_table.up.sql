CREATE TABLE IF NOT EXISTS commits (
                                       id SERIAL PRIMARY KEY,
                                       repository_id INT NOT NULL,
                                       hash VARCHAR(40) NOT NULL UNIQUE,
    message TEXT NOT NULL,
    author_name VARCHAR(255),
    author_email VARCHAR(255),
    commit_date TIMESTAMPTZ NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
    );

-- Index on repository_id for efficient joins and lookups
CREATE INDEX IF NOT EXISTS idx_commits_repository_id ON commits(repository_id);

-- Index on commit_date for efficient sorting and filtering by date
CREATE INDEX IF NOT EXISTS idx_commits_commit_date ON commits(commit_date);