# Variables
BINARY_SERVER=gophkeeper-server
BINARY_CLIENT=gophkeeper-client
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DATE=$(shell date +'%Y-%m-%dT%H:%M:%S')
LDFLAGS=-ldflags "-X main.buildVersion=$(VERSION) -X main.buildDate=$(DATE)"

.PHONY: all build-server build-client test test-coverage proto lint clean run-server

all: build-server build-client

# Build server
build-server:
	@echo "Building server..."
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_SERVER) cmd/server/main.go

# Build client
build-client:
	@echo "Building client..."
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_CLIENT) cmd/client/main.go

# Generate proto code
proto:
	@echo "Generating proto..."
	mkdir -p internal/proto/gophkeeperv1
	protoc --proto_path=proto --go_out=. --go_opt=module=github.com/tigranqic/gophkeeper-diploma \
		--go-grpc_out=. --go-grpc_opt=module=github.com/tigranqic/gophkeeper-diploma \
		proto/gophkeeper.proto

# Run all tests
test:
	@echo "Running tests..."
	go test ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out

# Run server (example with default args)
run-server: build-server
	./$(BUILD_DIR)/$(BINARY_SERVER) -a localhost:8080 -g localhost:3200

# Help command
help:
	@echo "Usage:"
	@echo "  make build-server   - Build server binary"
	@echo "  make build-client   - Build client binary"
	@echo "  make test           - Run all unit tests"
	@echo "  make test-coverage  - Run tests and show coverage"
	@echo "  make proto          - Generate gRPC code from .proto"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make clean          - Remove build artifacts"
