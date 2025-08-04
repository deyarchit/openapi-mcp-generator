SHELL = /bin/bash

# Project configs
APP_NAME=openapi-mcp-generator
BIN_DIR=bin
BUILD_DIR=$(BIN_DIR)/$(APP_NAME)

# Go commands
GO=go
GOFMT=gofmt
LINT=golangci-lint
GOBUILD=GOWORK=off $(GO) build -mod=vendor -o $(BUILD_DIR)
GOTEST=$(GO) test -v ./...

# Default target
all: build

# Build the binary
build:
	@echo "🔨 Building $(APP_NAME)..."
	$(GOBUILD) ./cmd/mcp-server-cli/main.go

# Run the server
run: build
	@echo "🚀 Running $(APP_NAME)..."
	@$(BUILD_DIR)

# Tidy dependencies
tidy:
	@echo "🧹 Cleaning up dependencies..."
	$(GO) mod tidy
	$(GO) mod vendor

# Run tests
test:
	@echo "✅ Running tests..."
	SKIP_INTEG=true $(GOTEST)

# Run case study loader
test-case-studies:
	@echo "✅ Loading case studies..."
	$(GO) test -v -count=1 -run TestPopulateCaseStudy ./...

# Lint the code
lint:
	@echo "🔍 Running linter..."
	$(LINT) run ./...

# Format code
fmt:
	@echo "🖊️ Formatting code..."
	$(GOFMT) -w .


# Prepare PR
pr: tidy fmt lint test
		@echo "🖊️ Formatting code..."
		$(GOFMT) -w .


# Clean build artifacts
clean:
	@echo "🗑️ Cleaning up..."
	rm -rf $(BUILD_DIR)

.PHONY: *
