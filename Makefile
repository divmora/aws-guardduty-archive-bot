.PHONY: build clean test test-coverage run-cli run-approve docker-build lambda-package dev-setup fmt lint help

# Build all binaries to bin/
build:
	@echo "Building binaries to bin/..."
	@mkdir -p bin
	@go build -o bin/aws-guardduty-archive-bot ./cmd/lambda
	@cp bin/aws-guardduty-archive-bot bin/lambda
	@echo "✓ Build complete! Binaries in ./bin/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f bootstrap lambda.zip
	@echo "✓ Clean complete!"

# Run tests
test:
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@go test -cover ./...

# Build Docker image
docker-build:
	@docker build --provenance=false --no-cache -t aws-guardduty-archive-bot:latest .

# Build Lambda deployment package
lambda-package: build
	@echo "Creating Lambda deployment package..."
	@cd bin && zip -r ../lambda.zip lambda
	@echo "✓ Lambda package created: lambda.zip"

# Install development dependencies
dev-setup:
	@go mod download
	@go install golang.org/x/tools/gopls@latest

# Format code
fmt:
	@go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@golangci-lint run

# Run in CLI dry-run mode
run-cli:
	@go run ./cmd/lambda

# Run in CLI approve mode
run-approve:
	@go run ./cmd/lambda -approve

help:
	@echo "Available targets:"
	@echo "  make build          - Build all binaries to bin/"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make run-cli        - Run locally in CLI dry-run mode"
	@echo "  make run-approve    - Run locally in CLI approve mode"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make lambda-package - Create Lambda deployment package"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Lint code"
