#!/bin/sh
set -e

./wait-for-it.sh postgres:5432 -- echo "PostgreSQL is up"

echo "Starting the application..."
exec ./goapp
