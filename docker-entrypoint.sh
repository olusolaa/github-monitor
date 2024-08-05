#!/bin/sh
set -e

if [ -f .env ]; then
  export $(cat .env | grep -v '#' | awk '/=/ {print $1}')
fi

./wait-for-it.sh postgres:5432 -- echo "PostgreSQL is up"

echo "Running migrations..."
migrate -path "migrations" -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:5432/${POSTGRES_DB}?sslmode=disable" up

echo "Starting the application..."
exec ./goapp
