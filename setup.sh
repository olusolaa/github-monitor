#!/bin/bash

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
  echo "START_DATE=2024-08-03T15:20:00Z"
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

docker-compose up -d