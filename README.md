
# GitHub Monitor

GitHub Monitor is a service designed to monitor GitHub repositories, track commits, and manage repository information using PostgreSQL and RabbitMQ. The project is implemented in Go and provides APIs for fetching repository data, commit history, and more.

## Table of Contents

- [Folder Structure](#folder-structure)
- [Installation](#installation)
- [Environment Variables](#environment-variables)
- [Makefile Commands](#makefile-commands)
- [API Routes](#api-routes)
- [Core Logic](#core-logic)
- [Contribution](#contribution)
- [License](#license)

## Folder Structure

```plaintext
cmd/
  └── main.go
config/
  └── config.go
db/
  └── migrations/
        ├── 20240730213851_create_repositories_table.down.sql
        ├── 20240730213851_create_repositories_table.up.sql
        ├── 20240730214425_create_commits_table.down.sql
        └── 20240730214425_create_commits_table.up.sql
internal/
  └── adapters/
        ├── cache/
        ├── consumers/
              ├── commit_consumer.go
              └── monitoring_consumer.go
        ├── github/
              ├── client.go
              └── types.go
        └── http/
              ├── handler.go
              └── middleware.go
  └── postgresdb/
        ├── commit.go
        └── repository.go
  └── queue/
        ├── connection_manager.go
        ├── message_consumer.go
        └── message_publisher.go
  └── container/
        └── container.go
  └── core/
        ├── domain/
              ├── commit.go
              └── repository.go
        ├── initializer/
              └── repository_initializer.go
        ├── services/
              ├── commit.go
              ├── github.go
              ├── monitor.go
              └── repository.go
        └── scheduler/
              └── scheduler.go
pkg/
  └── errors/
        └── errors.go
  └── httpclient/
        └── client.go
  └── logger/
        └── logger.go
  └── pagination/
        └── pagination.go
  └── utils/
        └── util.go
test/
  └── .env
go.mod
Makefile
```

## Installation

To set up the project locally, follow these steps:

1. **Clone the repository:**

    ```sh
    # Using SSH
    git clone git@github.com:olusolaa/github-monitor.git

    # Using HTTPS
    git clone https://github.com/olusolaa/github-monitor.git
    ```

2. **Install dependencies:**

    Ensure you have [Go](https://golang.org/doc/install) installed. Then, install other dependencies:

    ```sh
    go mod tidy
    ```

3. **Install PostgreSQL and RabbitMQ:**

    On macOS, you can install these using Homebrew:

    ```sh
    brew install postgresql
    brew install rabbitmq
    ```

4. **Set up environment variables:**

    Create a `.env` file in the root directory and populate it with the necessary environment variables (see the sample below).

## Environment Variables

Here's a sample `.env` file you can use to set up the project. Create a `.env` file in the root directory and populate it with the following environment variables:

```env
# Server settings
SERVER_ADDRESS=0.0.0.0:8080

# GitHub settings
GITHUB_TOKEN=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
GITHUB_BASE_URL=https://api.github.com

# Database settings
DATABASE_DSN=postgresql://localhost:5432/github_monitor_db?sslmode=disable

# Polling and webhook settings
POLL_INTERVAL=3600
MAX_RETRIES=3
INITIAL_BACKOFF=2
START_DATE=2024-08-03T15:20:00Z
END_DATE=2024-08-03T15:30:00Z
WEBHOOK_SECRET=your_webhook_secret_here

# Other settings
LOG_LEVEL=info

DEFAULT_OWNER=chromium
DEFAULT_REPO=chromium
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

## Makefile Commands

The Makefile includes various commands for managing the project. Key commands include:

- **Install migration tool for Mac:**
    ```sh
    make install-migrate-mac
    ```

- **Run migrations:**
    ```sh
    make migrate-up-all
    ```

- **Run the application:**
    ```sh
    make run
    ```

- **Run tests:**
    ```sh
    make test
    ```

## API Routes

The following routes are available in the application:

- **GET /api/repos/{owner}/{repo}** - Get repository details.
- **GET /api/repos/{owner}/{repo}/commits** - Get commits for a repository.
- **GET /api/repos/{repo_id}/top-authors** - Get top authors by commit count.
- **POST /api/repos/{repo_id}/reset-collection** - Reset the collection of a repository.

## Core Logic

The core logic of the application is primarily located in the `internal` and `internal/core/services` directories. The `services` package contains business logic related to repositories, commits, GitHub interactions, and monitoring.