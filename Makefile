.PHONY: build test clean install dist help

# Binary name
BINARY=signatured
VERSION=$(shell cat VERSION 2>/dev/null || echo "dev")

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary for current platform
	@echo "Building $(BINARY)..."
	@go build $(LDFLAGS) -o $(BINARY) ./cmd/signatured
	@echo "Build complete: ./$(BINARY)"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY)
	@rm -rf dist/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

install: build ## Install binary to /usr/local/bin
	@echo "Installing $(BINARY) to /usr/local/bin..."
	@sudo cp $(BINARY) /usr/local/bin/
	@echo "Installation complete"

dist: clean ## Build binaries for all platforms
	@echo "Building for all platforms..."
	@mkdir -p dist
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 ./cmd/signatured
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 ./cmd/signatured
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 ./cmd/signatured
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe ./cmd/signatured
	@echo "Distribution binaries built in dist/"
	@ls -lh dist/

validate-template: build ## Validate the signature template
	@./$(BINARY) validate

dry-run: build ## Run dry-run against test user (requires TEST_USER and ADMIN_USER env vars)
	@./$(BINARY) apply --user $(TEST_USER) --impersonate $(ADMIN_USER) --dry-run

lint: ## Run linters
	@echo "Running linters..."
	@go vet ./...
	@gofmt -l .

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

mod-tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@go mod tidy

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download

run-tests-verbose: ## Run tests with verbose output
	@go test -v -race ./...
