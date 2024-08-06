
# GitHub Monitor

GitHub Monitor is a service designed to monitor GitHub repositories, track commits, and manage repository information using PostgreSQL and RabbitMQ. The project is implemented in Go and provides APIs for fetching repository data, commit history, and more.

## Table of Contents

- [Folder Structure](#folder-structure)
- [Installation](#installation)
- [API Routes](#api-routes)
- [Core Logic](#core-logic)

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
        ├── github/
              ├── client.go
              ├── pagination.go
              ├── ratelimiter.go
              └── types.go
        ├── http/
              ├── handler.go
              └── middleware.go
        ├── postgresdb/
              ├── commit.go
              └── repository.go
        ├── queue/
              ├── connection_manager.go
              ├── message_consumer.go
              └── message_publisher.go
  └── container/
        └── container.go
  └── core/
        ├── domain/
              ├── commit.go
              └── repository.go
        ├── services/
              ├── commit.go
              ├── github.go
              ├── monitor.go
              └── repository.go
        ├── scheduler/
              └── scheduler.go
pkg/
  └── errors/
        └── errors.go
  └── httpclient/
        ├── client.go
        ├── middleware.go
        ├── request.go
        └── response.go
  └── logger/
        └── logger.go
  └── pagination/
        └── pagination.go
  └── utils/
        └── util.go
test/
  ├── commit_test.go
  └── repository_test.go
.env
docker-compose.yml
docker-entrypoint.sh
Dockerfile
go.mod
Makefile
README.md
setup.sh
wait-for-it.sh
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

3. **Run the setup script:** The .setup.sh script will automate the setup process. Run the following command:

    ```sh
    ./setup.sh
    ```

- **Note**: If you encounter permission issues, you may need to grant execute permissions to the script. You can do this by running:

    ```sh
    chmod +x setup.sh
    ```
  
During the script execution, you will encounter the following steps:

- **Docker Installation and Startup**:
  The script checks for Docker. If Docker is not installed, it will install Docker (on Linux) or prompt you to install Docker Desktop (on Mac and Windows). It then ensures Docker is running.
- **Environment Variable Setup**:
  The script sets up necessary environment variables. During this process, you will be prompted for:
    - **GitHub Token** `GITHUB_TOKEN`: **Required**. Enter a valid GitHub token for API access. The script will not continue without it.
    - **Optional Configurations**:
         - Start Date `START_DATE`: When to start pulling commits. Format: `YYYY-MM-DD` or `YYYY-MM-DDTHH:MM:SSZ`
         - End Date `END_DATE`: When to stop pulling commits. Same format as `START_DATE`.
         - Repository Owner `DEFAULT_OWNER`: The GitHub username of the repository owner.
         - Repository Name `DEFAULT_REPO`: The name of the repository to monitor.
         - Poll Interval `POLL_INTERVAL`: Interval in seconds to check for new commits.
- **Starting Docker Containers:**:
  The script will build and start Docker containers for the application and PostgreSQL.


## API Routes

The following routes are available in the application:

- **GET /api/repos/{owner}/{repo}** - Get repository details.
- **GET /api/repos/{owner}/{repo}/commits** - Get commits for a repository.
- **GET /api/repos/{owner}/{name}/top-authors** - Get top authors by commit count.
- **POST /api/repos/{owner}/{name}/reset-collection** - Reset the collection of a repository.
- **POST /api/repos/{owner}/{name}/monitor** - Add a new repository to the monitoring list.

## Core Logic

The core logic of the application is primarily located in the `internal` and `internal/core/services` directories. The `services` package contains business logic related to repositories, commits, GitHub interactions, and monitoring.

## Sample API Requests and Responses

#### 1. Fetch Repository Details
**Method**: GET  
**URL**: `http://localhost:8080/api/repos/chromium/chromium`  
**Sample Response**:
```json
{
    "name": "chromium",
    "description": "The official GitHub mirror of the Chromium source",
    "url": "https://api.github.com/repos/chromium/chromium",
    "language": "C++",
    "forks_count": 6805,
    "stargazers_count": 18393,
    "open_issues_count": 93,
    "watchers_count": 18393,
    "created_at": "2018-02-05T20:55:32Z",
    "updated_at": "2024-08-04T03:16:04Z"
}
```

#### 2. Fetch Commits for a Repository
**Method**: GET  
**URL**: `http://localhost:8080/api/repos/chromium/chromium/commits?page=1&page_size=50`  
**Sample Response**:
```json
{
    "pagination": {
        "page": 1,
        "page_size": 50,
        "total_pages": 42,
        "total_items": 2058
    },
    "data": [
        {
            "hash": "1d09fe5035f03602eadd372519b3e57c87701d6c",
            "message": "[iOSKAUpgrade] Fix overflow menu position for addresses without name",
            "author_name": "Chromium LUCI CQ",
            "author_email": "chromium-scoped@luci-project-accounts.iam.gserviceaccount.com",
            "commit_date": "2024-08-06T15:20:02+01:00",
            "url": "https://api.github.com/repos/chromium/chromium/git/commits/1d09fe5035f03602eadd372519b3e57c87701d6c"
        },
        {
            "hash": "cb97b87fe0b5e404d0b7b0c7eaec4e5baaf3c407",
            "message": "Create DTCKeyRotationUploadedBySharedAPIEnabled feature.",
            "author_name": "Chromium LUCI CQ",
            "author_email": "chromium-scoped@luci-project-accounts.iam.gserviceaccount.com",
            "commit_date": "2024-08-06T15:18:30+01:00",
            "url": "https://api.github.com/repos/chromium/chromium/git/commits/cb97b87fe0b5e404d0b7b0c7eaec4e5baaf3c407"
        }
    ]
}
```

#### 3. Reset Collection
**Method**: POST  
**URL**: `http://localhost:8080/api/repos/chromium/chromium/reset-collection?start_time=2024-08-01T00:00:00Z`  
**Sample Response**:
```json
{
    "message": "Collection reset successfully"
}
```

#### 4. Get Top Authors
**Method**: GET  
**URL**: `http://localhost:8080/api/repos/chromium/chromium/top-authors?limit=3`  
**Sample Response**:
```json
[
    {
        "author_name": "Chromium LUCI CQ",
        "author_email": "chromium-scoped@luci-project-accounts.iam.gserviceaccount.com",
        "commit_count": 2039
    },
    {
        "author_name": "Chrome Release Bot (LUCI)",
        "author_email": "chrome-official-brancher@chops-service-accounts.iam.gserviceaccount.com",
        "commit_count": 11
    },
    {
        "author_name": "Nico Weber",
        "author_email": "thakis@chromium.org",
        "commit_count": 4
    }
]
```

#### 5. Monitor a Repository
**Method**: POST  
**URL**: `http://localhost:8080/api/repos/facebook/react/monitor`  
**Sample Response**:
```json
{
    "message": "Repository monitoring triggered successfully"
}
```

These examples showcase how to interact with the GitHub Monitor API. The endpoints support various actions such as fetching repository details, commit history, resetting data collections, and starting monitoring for repositories.
