.PHONY: build test lint run clean install help

# Binary name
BINARY_NAME=workflow-core

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOFMT=$(GOCMD) fmt
GOMOD=$(GOCMD) mod

# Build flags
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

# Default target
all: build

## build: Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/workflow-core/

## build-all: Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 ./cmd/workflow-core/
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 ./cmd/workflow-core/
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 ./cmd/workflow-core/
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 ./cmd/workflow-core/
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe ./cmd/workflow-core/

## test: Run tests
test:
	$(GOTEST) -v -race -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	$(GOVET) ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, running go vet only"; \
	fi

## fmt: Format code
fmt:
	$(GOFMT) ./...

## run: Run the example workflow
run: build
	./$(BINARY_NAME) run examples/simple-workflow.json

## run-verbose: Run with verbose output
run-verbose: build
	./$(BINARY_NAME) run -verbose examples/simple-workflow.json

## validate: Validate example workflow
validate: build
	./$(BINARY_NAME) validate examples/simple-workflow.json

## list-nodes: List supported nodes
list-nodes: build
	./$(BINARY_NAME) list-nodes

## clean: Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out coverage.html

## deps: Download dependencies
deps:
	$(GOMOD) download

## tidy: Tidy go.mod
tidy:
	$(GOMOD) tidy

## install: Install the binary
install: build
	cp $(BINARY_NAME) $(GOPATH)/bin/

## version: Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"

## help: Show this help
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
