# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Copy the .env file, ensuring it is included in the build
COPY .env .env

# Build the application with CGO disabled for a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /goapp cmd/main.go

# Stage 2: Start from a minimal image
FROM alpine:latest
WORKDIR /root/

# Install necessary runtime dependencies, including bash, curl, and ca-certificates
RUN apk --no-cache add bash ca-certificates curl

# Copy the binary, the .env file, and wait-for-it.sh from the builder stage
COPY --from=builder /goapp .
COPY --from=builder /app/.env .env
COPY --from=builder /app/wait-for-it.sh ./

# Install gomigrate
ENV MIGRATE_VERSION v4.14.1
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate.linux-amd64 /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Make wait-for-it.sh executable
RUN chmod +x wait-for-it.sh

# Copy the migrations directory
COPY --from=builder /app/db/migrations /root/db/migrations

# Copy and set the entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Expose the port the application runs on
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
