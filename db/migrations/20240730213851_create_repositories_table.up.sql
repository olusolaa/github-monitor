CREATE TABLE IF NOT EXISTS repositories (
                                            id SERIAL PRIMARY KEY,
                                            name TEXT NOT NULL,
                                            owner TEXT NOT NULL,
                                            description TEXT,
                                            url TEXT NOT NULL,
                                            language TEXT,
                                            forks_count INT,
                                            stargazers_count INT,
                                            open_issues_count INT,
                                            watchers_count INT,
                                            created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(name, owner)
    );
