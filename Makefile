# =================================================================
# MS-AI Backend Makefile
# =================================================================

.PHONY: help build build-prod run dev test test-unit test-integration test-coverage lint fmt security check clean
.PHONY: docker-build docker-up docker-down docker-logs docker-clean
.PHONY: db-test db-health db-seed db-migrate
.PHONY: install-tools update-deps audit-deps
.PHONY: gen-mocks gen-docs gen-proto
.PHONY: ci cd deploy

# Variables
BUILD_BIN := bin/server
BUILD_TIME := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo 'dev')

# Build flags
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.buildTime=$(BUILD_TIME) \
	-X main.gitCommit=$(GIT_COMMIT)

# =================================================================
# DEVELOPMENT COMMANDS
# =================================================================

## help: Show this help message
help:
	@echo "MS-AI Backend - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

## build: Build the application for development
build:
	@echo "🔨 Building application..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS)" \
		-o $(BUILD_BIN) \
		./cmd/server
	@echo "✅ Build complete: $(BUILD_BIN)"

## build-prod: Build optimized production binary
build-prod:
	@echo "🏭 Building production binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags="$(LDFLAGS) -extldflags '-static'" \
		-tags netgo \
		-trimpath \
		-o $(BUILD_BIN) \
		./cmd/server
	@echo "✅ Production build complete: $(BUILD_BIN)"
	@du -h $(BUILD_BIN)

## run: Run the application
run: build
	@echo "🚀 Starting server..."
	./$(BUILD_BIN)

## dev: Run in development mode with hot reload
dev:
	@echo "🔄 Starting development server..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular run..."; \
		$(MAKE) run; \
	fi

# =================================================================
# TESTING COMMANDS
# =================================================================

## test: Run all tests
test:
	@echo "🧪 Running all tests..."
	go test ./... -v -race -timeout 30s

## test-unit: Run unit tests only
test-unit:
	@echo "🧪 Running unit tests..."
	go test ./... -v -race -timeout 30s -short

## test-integration: Run integration tests
test-integration:
	@echo "🔗 Running integration tests..."
	@if [ -z "$$MONGO_URI" ]; then \
		echo "❌ MONGO_URI not set. Run: export MONGO_URI=mongodb://localhost:27017"; \
		exit 1; \
	fi
	go test ./test/integration/... -v -race -timeout 60s

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "📊 Running tests with coverage..."
	go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "📈 Coverage report generated: coverage.html"

# =================================================================
# CODE QUALITY COMMANDS
# =================================================================

## lint: Run linter
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Install with: make install-tools"; \
	fi

## fmt: Format code
fmt:
	@echo "💅 Formatting code..."
	gofmt -w -s .
	goimports -w .

## security: Run security checks
security:
	@echo "🔒 Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: make install-tools"; \
	fi

## check: Run all code quality checks
check: fmt lint security test-unit
	@echo "✅ All checks passed!"

# =================================================================
# DOCKER COMMANDS
# =================================================================

## docker-build: Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t ms-ai-backend:$(VERSION) .
	docker tag ms-ai-backend:$(VERSION) ms-ai-backend:latest

## docker-up: Start all services with Docker Compose
docker-up:
	@echo "🚀 Starting services with Docker Compose..."
	docker-compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	$(MAKE) docker-health

## docker-down: Stop all services
docker-down:
	@echo "🛑 Stopping services..."
	docker-compose down

## docker-logs: Show service logs
docker-logs:
	docker-compose logs -f app

## docker-health: Check service health
docker-health:
	@echo "🏥 Checking service health..."
	@docker-compose ps
	@echo ""
	@echo "Application health check:"
	@curl -s http://localhost:8080/health | jq . || echo "❌ Health check failed"

## docker-clean: Clean up Docker resources
docker-clean:
	@echo "🧹 Cleaning up Docker resources..."
	docker-compose down -v --remove-orphans
	docker system prune -f

# =================================================================
# DATABASE COMMANDS
# =================================================================

## db-test: Test database connection
db-test:
	@echo "🗄️ Testing database connection..."
	@if [ -z "$$MONGO_URI" ]; then \
		echo "❌ MONGO_URI not set"; \
		exit 1; \
	fi
	go run -tags=dbtest ./scripts/dbtest.go

## db-health: Check database health
db-health:
	@echo "💚 Checking database health..."
	@docker-compose exec mongodb mongosh --eval "db.runCommand({ping:1})" --quiet || echo "❌ Database health check failed"

## db-seed: Seed database with test data
db-seed:
	@echo "🌱 Seeding database..."
	go run ./scripts/seed.go

## db-migrate: Run database migrations
db-migrate:
	@echo "📈 Running database migrations..."
	go run ./cmd/migrations/migrate.go

# =================================================================
# DEVELOPMENT TOOLS
# =================================================================

## install-tools: Install development tools
install-tools:
	@echo "🔧 Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/golang/mock/mockgen@latest

## update-deps: Update Go dependencies
update-deps:
	@echo "📦 Updating dependencies..."
	go mod tidy
	go mod download

## audit-deps: Audit dependencies for security issues
audit-deps:
	@echo "🔍 Auditing dependencies..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

## gen-mocks: Generate mocks for testing
gen-mocks:
	@echo "🤖 Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/core/auth/service.go -destination=test/mocks/auth_service_mock.go -package=mocks; \
		mockgen -source=internal/data/user/repository.go -destination=test/mocks/user_repository_mock.go -package=mocks; \
	else \
		echo "mockgen not found. Install with: make install-tools"; \
	fi

## gen-docs: Generate API documentation
gen-docs:
	@echo "📚 Generating API documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/server/main.go -o docs/swagger; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# =================================================================
# CI/CD COMMANDS
# =================================================================

## ci: Run CI pipeline locally
ci: check test-integration docker-build
	@echo "✅ CI pipeline completed successfully"

## cd: Run CD pipeline locally (build and deploy)
cd: build-prod docker-build
	@echo "🚀 CD pipeline completed - ready for deployment"

## deploy: Deploy to production (customize for your environment)
deploy:
	@echo "🚀 Deploying to production..."
	@echo "Customize this target for your deployment strategy"
	@echo "Examples: Kubernetes, Docker Swarm, AWS ECS, etc."

# =================================================================
# UTILITY COMMANDS
# =================================================================

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning up..."
	rm -rf bin/ coverage.out coverage.html
	go clean ./...

## info: Show project information
info:
	@echo "📋 MS-AI Backend Information:"
	@echo "  Version: $(VERSION)"
	@echo "  Build Time: $(BUILD_TIME)"
	@echo "  Git Commit: $(GIT_COMMIT)"
	@echo "  Go Version: $(shell go version)"
	@echo "  Build Target: $(BUILD_BIN)"

# Default target
.DEFAULT_GOAL := help

db-migrate:
	@echo "=> No automated migrations configured. Add migration tooling (e.g. migrate) and implement this target."

db-seed: seed

db-stats:
	@echo "=> Attempting to show MongoDB stats (requires mongo shell)"
	@mongo $(MONGO_URI) --quiet --eval 'db.adminCommand({serverStatus:1})' || echo "mongo shell not available"

db-health:
	@echo "=> Checking application /health endpoint"
	@curl -fsS http://localhost:8080/health || echo "health endpoint not reachable"

test-repos:
	@echo "=> Running repository tests"
	@go test ./internal/data/... -v
