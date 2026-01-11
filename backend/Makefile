.PHONY: build run test lint fmt clean dev

# Build the application
build:
	go build -o bin/api ./cmd/api

# Run the application
run: build
	./bin/api

# Run in development mode with hot reload (requires air)
dev:
	air

# Run tests
test:
	go test -v -cover ./...

# Run tests with race detector
test-race:
	go test -race -v ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Download dependencies
deps:
	go mod download
	go mod tidy

# Generate (placeholder for future code generation)
generate:
	go generate ./...

# Setup development environment
setup:
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Copying .env.example to .env..."
	@if [ ! -f .env ]; then cp .env.example .env; fi
	@echo "Setup complete!"

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  run        - Build and run the application"
	@echo "  dev        - Run with hot reload (requires air)"
	@echo "  test       - Run tests"
	@echo "  test-race  - Run tests with race detector"
	@echo "  lint       - Run linter"
	@echo "  fmt        - Format code"
	@echo "  clean      - Clean build artifacts"
	@echo "  deps       - Download and tidy dependencies"
	@echo "  setup      - Setup development environment"
