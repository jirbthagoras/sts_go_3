.PHONY: help build run test clean docker-up docker-down swagger

# Default target
help:
	@echo "Available commands:"
	@echo "  build       - Build the application"
	@echo "  run         - Run the application"
	@echo "  test        - Run tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  docker-up   - Start PostgreSQL with Docker Compose"
	@echo "  docker-down - Stop PostgreSQL Docker containers"
	@echo "  swagger     - Generate Swagger documentation"
	@echo "  deps        - Download dependencies"

# Build the application
build:
	go build -o bin/film-api .

# Run the application
run:
	go run .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -rf docs/

# Start PostgreSQL with Docker Compose
docker-up:
	docker-compose up -d postgres

# Stop PostgreSQL Docker containers
docker-down:
	docker-compose down

# Generate Swagger documentation
swagger:
	swag init -g main.go -o docs/

# Download dependencies
deps:
	go mod tidy
	go mod download
