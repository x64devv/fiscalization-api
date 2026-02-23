.PHONY: help build run test clean migrate-up migrate-down docker-build docker-up docker-down

# Variables
APP_NAME=fiscalization-api
DOCKER_IMAGE=$(APP_NAME):latest
MAIN_PATH=cmd/server/main.go

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	go run $(MAIN_PATH)

test: ## Run tests
	@echo "Running tests..."
	go test -v -cover ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration $(NAME)..."
	migrate create -ext sql -dir migrations -seq $(NAME)

migrate-up: ## Run database migrations up
	@echo "Running migrations up..."
	migrate -path migrations -database "postgresql://fiscalization:password@localhost:5432/fiscalization_db?sslmode=disable" up

migrate-down: ## Run database migrations down
	@echo "Running migrations down..."
	migrate -path migrations -database "postgresql://fiscalization:password@localhost:5432/fiscalization_db?sslmode=disable" down

migrate-force: ## Force migration version (usage: make migrate-force VERSION=1)
	@echo "Forcing migration to version $(VERSION)..."
	migrate -path migrations -database "postgresql://fiscalization:password@localhost:5432/fiscalization_db?sslmode=disable" force $(VERSION)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start Docker containers
	@echo "Starting Docker containers..."
	docker-compose up -d

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

security: ## Run security checks
	@echo "Running security checks..."
	gosec ./...

generate: ## Generate code (mocks, etc.)
	@echo "Generating code..."
	go generate ./...

seed: ## Seed database with test data
	@echo "Seeding database..."
	go run scripts/seed.go

install-tools: ## Install development tools
	@echo "Installing tools..."
	go install github.com/golang/mock/mockgen@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.DEFAULT_GOAL := help
