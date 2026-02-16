BINARY_NAME=gitops-time-machine
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: all build test clean lint fmt docker help

all: fmt lint test build ## Run all checks and build

build: ## Build the binary
	@echo "🔨 Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

test: ## Run unit tests
	@echo "🧪 Running tests..."
	go test ./... -v -cover -coverprofile=coverage.out

test-short: ## Run short tests only
	@echo "🧪 Running short tests..."
	go test ./... -short -v

coverage: test ## Generate coverage report
	@echo "📊 Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	@echo "🔍 Linting..."
	@which golangci-lint > /dev/null 2>&1 || echo "Install golangci-lint: https://golangci-lint.run/usage/install/"
	golangci-lint run ./...

fmt: ## Format code
	@echo "✨ Formatting..."
	go fmt ./...

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf bin/ coverage.out coverage.html

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run: ## Run in Docker
	docker run --rm -v ~/.kube:/root/.kube $(BINARY_NAME):$(VERSION)

install: build ## Install binary to GOPATH
	@echo "📦 Installing..."
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
