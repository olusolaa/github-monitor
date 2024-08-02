#!/bin/bash

# Set variables
REPO_HTTPS_URL="https://github.com/olusolaa/github-monitor.git"
ENV_FILE=".env"
DB_NAME="github_monitor_db"
NON_EXISTENT_MIGRATION="20240731083505"

# Clone the repository
echo "Cloning the repository..."
git clone $REPO_HTTPS_URL
cd github-monitor || exit

# Install Go if not installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Installing Go..."
    brew install go
else
    echo "Go is already installed."
fi

# Install Go dependencies
echo "Installing Go dependencies..."
go mod tidy

# Install PostgreSQL if not installed
if ! command -v psql &> /dev/null; then
    echo "PostgreSQL is not installed. Installing PostgreSQL..."
    brew install postgresql
    echo "Starting PostgreSQL service..."
    brew services start postgresql
else
    echo "PostgreSQL is already installed."
    echo "Starting PostgreSQL service..."
    brew services start postgresql
fi

# Install RabbitMQ if not installed
if ! command -v rabbitmqctl &> /dev/null; then
    echo "RabbitMQ is not installed. Installing RabbitMQ..."
    brew install rabbitmq
    echo "Starting RabbitMQ service..."
    brew services start rabbitmq
else
    echo "RabbitMQ is already installed."
    echo "Starting RabbitMQ service..."
    brew services start rabbitmq
fi

# Create PostgreSQL role if it doesn't exist
echo "Checking if PostgreSQL role 'postgres' exists..."
ROLE_EXIST=$(psql -U postgres -tc "SELECT 1 FROM pg_roles WHERE rolname = 'postgres';")
if [ "$ROLE_EXIST" != "1" ]; then
    echo "Creating PostgreSQL role 'postgres'..."
    sudo -u postgres createuser --superuser postgres
fi

# Create PostgreSQL database if it doesn't exist
echo "Checking if database '$DB_NAME' exists..."
DB_EXIST=$(psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME';")
if [ "$DB_EXIST" == "1" ]; then
    echo "Database '$DB_NAME' already exists."
else
    echo "Creating PostgreSQL database '$DB_NAME'..."
    psql -U postgres -c "CREATE DATABASE $DB_NAME;"
fi

# Create .env file
echo "Setting up environment variables..."
cat <<EOT > $ENV_FILE
# Server settings
SERVER_ADDRESS=0.0.0.0:8080

# GitHub settings
GITHUB_TOKEN=your_github_token_here

# Database settings
DATABASE_DSN=postgresql://localhost:5432/$DB_NAME?sslmode=disable

# Polling and webhook settings
POLL_INTERVAL=3600
MAX_RETRIES=3
INITIAL_BACKOFF=2
START_DATE=2023-01-01T00:00:00Z
WEBHOOK_SECRET=your_webhook_secret_here

# Other settings
LOG_LEVEL=info

DEFAULT_OWNER=chromium
DEFAULT_REPO=chromium
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
EOT

echo ".env file created."

# Install migration tool if not installed
if ! command -v migrate &> /dev/null; then
    echo "Installing migration tool for Mac..."
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.darwin-amd64.tar.gz | tar xvz
    mv migrate.darwin-amd64 /usr/local/bin/migrate
    chmod +x /usr/local/bin/migrate
else
    echo "Migration tool is already installed."
fi

# Remove non-existent migration reference (if applicable)
echo "Checking for non-existent migration references..."
MIGRATION_EXIST=$(psql -U postgres -d $DB_NAME -tc "SELECT version FROM schema_migrations WHERE version = '$NON_EXISTENT_MIGRATION';")
if [ "$MIGRATION_EXIST" == "$NON_EXISTENT_MIGRATION" ]; then
    echo "Removing reference to non-existent migration version $NON_EXISTENT_MIGRATION..."
    psql -U postgres -d $DB_NAME -c "DELETE FROM schema_migrations WHERE version = '$NON_EXISTENT_MIGRATION';"
else
    echo "No reference to non-existent migration version $NON_EXISTENT_MIGRATION found."
fi

# Run migrations
echo "Running migrations..."
if /usr/local/bin/migrate -path db/migrations -database "postgresql://localhost:5432/$DB_NAME?sslmode=disable" up; then
    echo "Migrations ran successfully."
else
    echo "Error running migrations. Check the migration files and database setup."
    exit 1
fi

# Build and run the application
echo "Building and running the application..."
if go build -o bin/main cmd/main.go; then
    echo "Build successful."
    ./bin/main
else
    echo "Build failed. Check for errors."
    exit 1
fi

echo "Setup completed successfully."
