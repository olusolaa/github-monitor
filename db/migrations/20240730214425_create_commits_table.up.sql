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
