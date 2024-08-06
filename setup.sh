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

# Wait for Docker to be ready
echo "Waiting for Docker to start..."
while ! docker info &> /dev/null; do
    echo -n "."
    sleep 1
done
echo "Docker is ready."

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

# Prompt user for GitHub token
echo "Please enter your GitHub token:"
read -r GITHUB_TOKEN

if [ -z "$GITHUB_TOKEN" ] || [ "$GITHUB_TOKEN" == "enter-your-github-token-here" ]; then
    echo "Error: A valid GitHub token is required."
    exit 1
fi

# Function to convert date to ISO 8601 format if necessary
convert_to_iso8601() {
    local input_date="$1"
    if [[ "$input_date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]]; then
        echo "${input_date}T00:00:00Z"
    else
        echo "$input_date"
    fi
}

# Function to validate date format
validate_date_format() {
    local date="$1"
    if [[ "$date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}$ ]] || [[ "$date" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$ ]]; then
        return 0
    else
        return 1
    fi
}

# Prompt for configurable environment variables with defaults

# START_DATE
DEFAULT_START_DATE="2024-08-03T01:20:00Z"
echo "Current start date for pulling commits is set to $DEFAULT_START_DATE."
while true; do
    echo "Enter a new start date in the format YYYY-MM-DD or YYYY-MM-DDTHH:MM:SSZ or press Enter to keep the default:"
    read -r START_DATE_INPUT
    START_DATE_INPUT="${START_DATE_INPUT:-$DEFAULT_START_DATE}"
    if validate_date_format "$START_DATE_INPUT"; then
        START_DATE=$(convert_to_iso8601 "$START_DATE_INPUT")
        break
    else
        echo "Invalid date format. Please enter the date in the correct format."
    fi
done

# END_DATE
DEFAULT_END_DATE="2024-08-03T15:30:00Z"
echo "Current end date for pulling commits is set to $DEFAULT_END_DATE."
while true; do
    echo "Enter a new end date in the format YYYY-MM-DD or YYYY-MM-DDTHH:MM:SSZ or press Enter to keep the default:"
    read -r END_DATE_INPUT
    END_DATE_INPUT="${END_DATE_INPUT:-$DEFAULT_END_DATE}"
    if validate_date_format "$END_DATE_INPUT"; then
        END_DATE=$(convert_to_iso8601 "$END_DATE_INPUT")
        break
    else
        echo "Invalid date format. Please enter the date in the correct format."
    fi
done

# DEFAULT_OWNER
DEFAULT_OWNER="chromium"
echo "Current default owner (GitHub username) is set to $DEFAULT_OWNER."
echo "Enter a new GitHub username for the repo owner or press Enter to keep the default:"
read -r OWNER
OWNER="${OWNER:-$DEFAULT_OWNER}"

# DEFAULT_REPO
DEFAULT_REPO="chromium"
echo "Current default repository name is set to $DEFAULT_REPO."
echo "Enter a new repository name or press Enter to keep the default:"
read -r REPO
REPO="${REPO:-$DEFAULT_REPO}"

# POLL_INTERVAL
DEFAULT_POLL_INTERVAL=3600
echo "Current poll interval (in seconds) for monitoring repository changes is set to $DEFAULT_POLL_INTERVAL."
echo "Enter a new poll interval in seconds or press Enter to keep the default:"
read -r POLL_INTERVAL
POLL_INTERVAL="${POLL_INTERVAL:-$DEFAULT_POLL_INTERVAL}"

# Add environment variables to .env
echo "Adding environment variables to .env"
{
  echo "SERVER_ADDRESS=0.0.0.0:8080"
  echo "GITHUB_TOKEN=$GITHUB_TOKEN"
  echo "GITHUB_BASE_URL=https://api.github.com"
  echo "DATABASE_DSN=postgresql://localhost:5432/github_monitor_db?sslmode=disable"
  echo "POLL_INTERVAL=$POLL_INTERVAL"
  echo "MAX_RETRIES=3"
  echo "INITIAL_BACKOFF=2"
  echo "START_DATE=$START_DATE"
  echo "END_DATE=$END_DATE"
  echo "LOG_LEVEL=info"
  echo "DEFAULT_OWNER=$OWNER"
  echo "DEFAULT_REPO=$REPO"
  echo "POSTGRES_USER=postgres"
  echo "POSTGRES_PASSWORD=postgres"
  echo "POSTGRES_DB=postgres"
  echo "POSTGRES_HOST=postgres"
} > .env

# Start Docker containers
docker-compose up