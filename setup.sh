#!/bin/bash

# Detect OS
OS=$(uname -s)
case "$OS" in
    Linux*)
        OS=Linux
        ;;
    Darwin*)
        OS=Mac
        ;;
    CYGWIN*|MINGW32*|MSYS*|MINGW*)
        OS=Windows
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# Check and install Docker if not installed
if ! command -v docker &> /dev/null; then
    echo "Docker not found, installing..."
    if [ "$OS" == "Linux" ]; then
        curl -fsSL https://get.docker.com -o get-docker.sh
        sh get-docker.sh
        rm get-docker.sh
    elif [ "$OS" == "Mac" ]; then
        echo "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
        exit 1
    elif [ "$OS" == "Windows" ]; then
        echo "Please install Docker Desktop from https://www.docker.com/products/docker-desktop"
        exit 1
    fi
else
    echo "Docker is already installed."
fi

# Start Docker if not running
if [ "$OS" == "Linux" ]; then
    if ! systemctl is-active --quiet docker; then
        echo "Starting Docker..."
        sudo systemctl start docker
        sudo systemctl enable docker
    else
        echo "Docker is already running."
    fi
elif [ "$OS" == "Mac" ]; then
    echo "Starting Docker Desktop on Mac."
    open --background -a Docker
elif [ "$OS" == "Windows" ]; then
    echo "Please ensure Docker Desktop is running."
fi

# Check and install Go if not installed
if ! command -v go &> /dev/null; then
    echo "Go not found, installing..."
    if [ "$OS" == "Linux" ]; then
        curl -LO https://golang.org/dl/go1.20.4.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.20.4.linux-amd64.tar.gz
        rm go1.20.4.linux-amd64.tar.gz
        export PATH=$PATH:/usr/local/go/bin
    elif [ "$OS" == "Mac" ]; then
        brew install go
    elif [ "$OS" == "Windows" ]; then
        echo "Please install Go from https://golang.org/dl/"
        exit 1
    fi
else
    echo "Go is already installed."
fi

# Add environment variables to .env
echo "Adding environment variables to .env"
{
  echo "SERVER_ADDRESS=0.0.0.0:8080"
  echo "GITHUB_TOKEN=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"
  echo "GITHUB_BASE_URL=https://api.github.com"
  echo "DATABASE_DSN=postgresql://localhost:5432/github_monitor_db?sslmode=disable"
  echo "POLL_INTERVAL=3600"
  echo "MAX_RETRIES=3"
  echo "INITIAL_BACKOFF=2"
  echo "START_DATE=2024-08-03T01:20:00Z"
  echo "END_DATE=2024-08-03T15:30:00Z"
  echo "WEBHOOK_SECRET=your_webhook_secret_here"
  echo "LOG_LEVEL=info"
  echo "DEFAULT_OWNER=chromium"
  echo "DEFAULT_REPO=chromium"
  echo "REDIS_URL=redis://localhost:6379"
  echo "RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/"
  echo "POSTGRES_USER=postgres"
  echo "POSTGRES_PASSWORD=postgres"
  echo "POSTGRES_DB=postgres"
  echo "POSTGRES_HOST=postgres"
} > .env

# Start Docker containers
docker-compose up