.PHONY: build run test test-unit test-integration clean docker-up docker-down docker-build help

# Build the application
build:
	go build -o bin/rate-limiter cmd/main.go

# Run the application locally
run:
	go run cmd/main.go

# Run all tests
test:
	go test ./... -v

# Run unit tests only
test-unit:
	go test ./pkg/ratelimiter -v

# Run integration tests only
test-integration:
	go test ./test -v

# Run tests with coverage
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f bin/rate-limiter
	rm -f coverage.out coverage.html

# Start services with Docker Compose
docker-up:
	docker-compose up -d

# Stop services
docker-down:
	docker-compose down

# Build Docker image
docker-build:
	docker build -t rate-limiter .

# Start only Redis
redis:
	docker-compose up -d redis

# View application logs
logs:
	docker-compose logs -f app

# View Redis logs
redis-logs:
	docker-compose logs -f redis

# Download dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/rakyll/hey@latest

# Quick test with curl (requires running app)
test-api:
	@echo "Testing basic endpoint..."
	curl -s http://localhost:8080/ | jq .
	@echo "\nTesting with token..."
	curl -s -H "API_KEY: abc123" http://localhost:8080/ | jq .

# Load test (requires hey and running app)
load-test:
	hey -n 50 -c 5 http://localhost:8080/

# Setup development environment
setup: deps install-tools
	cp env.example .env
	@echo "Development environment setup complete!"
	@echo "Edit .env file with your configuration and run 'make docker-up' to start"

# Help
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application locally"
	@echo "  test            - Run all tests"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration- Run integration tests only"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker-up       - Start services with Docker Compose"
	@echo "  docker-down     - Stop Docker Compose services"
	@echo "  docker-build    - Build Docker image"
	@echo "  redis           - Start only Redis service"
	@echo "  logs            - View application logs"
	@echo "  redis-logs      - View Redis logs"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  install-tools   - Install development tools"
	@echo "  test-api        - Quick API test"
	@echo "  load-test       - Load test with hey"
	@echo "  setup           - Setup development environment"
	@echo "  help            - Show this help message" 