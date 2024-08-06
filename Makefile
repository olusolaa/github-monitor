MIGRATE_VERSION=v4.14.1
MIGRATE_URL=https://github.com/golang-migrate/migrate/releases/download/$(MIGRATE_VERSION)
MIGRATE_MAC=migrate.darwin-amd64
MIGRATE_WINDOWS=migrate.windows-amd64.exe
MIGRATE_BIN=/usr/local/bin/migrate

DB_URL=postgresql://localhost:5432/postgres?sslmode=disable
MIGRATE_CMD=migrate -path db/migrations -database "$(DB_URL)"

# Install migration tool for Mac
install-migrate-mac:
	curl -L $(MIGRATE_URL)/$(MIGRATE_MAC).tar.gz | tar xvz
	mv $(MIGRATE_MAC) $(MIGRATE_BIN)
	chmod +x $(MIGRATE_BIN)

# Install migration tool for Windows
install-migrate-windows:
	curl -L $(MIGRATE_URL)/$(MIGRATE_WINDOWS).tar.gz -o $(MIGRATE_WINDOWS).tar.gz
	tar -xvzf $(MIGRATE_WINDOWS).tar.gz
	mv $(MIGRATE_WINDOWS) $(MIGRATE_BIN)

# Migration commands
migrate-up-all:
	$(MIGRATE_CMD) up

migrate-up:
	$(MIGRATE_CMD) up 1

migrate-down-all:
	$(MIGRATE_CMD) down

migrate-down:
	$(MIGRATE_CMD) down 1

migrate-force:
	@if [ -z "$(version)" ]; then echo "version is not set. Set it like this: make migrate-force version=4"; exit 1; fi
	$(MIGRATE_CMD) force $(version)

migration:
	@if [ -z "$(name)" ]; then echo "name is not set. Set it like this: make migration name=create_users"; exit 1; fi
	$(MIGRATE_BIN) create -ext sql -dir db/migrations -seq $(name)

# Local development commands
run:
	go run cmd/main.go

build:
	go build -o bin/main cmd/main.go

test:
	go test -v ./...