#!/bin/sh
set -e

# Load environment variables from .env file
if [ -f .env ]; then
  export $(cat .env | grep -v '#' | awk '/=/ {print $1}')
fi

# Run database migrations
echo "Running migrations..."
migrate -path "migrations" -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/${POSTGRES_DB}?sslmode=disable" up

# Run the application
echo "Starting the application..."
exec ./goapp
