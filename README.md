
# GitHub Monitor

GitHub Monitor is a service designed to monitor GitHub repositories, track commits, and manage repository information using PostgreSQL and RabbitMQ. The project is implemented in Go and provides APIs for fetching repository data, commit history, and more.

## Table of Contents

- [Folder Structure](#folder-structure)
- [Installation](#installation)
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

2. **Navigate to the project directory:**
    ```sh
   cd github-monitor
   ```

3. **Run the setup script:**

   Please update the ./setup.sh file with your github token and run the following command
    ```sh
    ./setup.sh
    ```
This script will:

- Check for Docker installation and start Docker if necessary.
- Set up environment variables from the .env file.
- Build and start the Docker containers for the application, PostgreSQL, and RabbitMQ.



## API Routes

The following routes are available in the application:

- **GET /api/repos/{owner}/{repo}** - Get repository details.
- **GET /api/repos/{owner}/{repo}/commits** - Get commits for a repository.
- **GET /api/repos/{owner}/{name}/top-authors** - Get top authors by commit count.
- **POST /api/repos/{owner}/{name}/reset-collection** - Reset the collection of a repository.
- **POST /api/repos/{owner}/{name}/monitor** - Add a new repository to the monitoring list.

## Core Logic

The core logic of the application is primarily located in the `internal` and `internal/core/services` directories. The `services` package contains business logic related to repositories, commits, GitHub interactions, and monitoring.