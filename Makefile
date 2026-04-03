# Variables
BINARY_SERVER=gophkeeper-server
BINARY_CLIENT=gophkeeper-client
BUILD_DIR=bin
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DATE=$(shell date +'%Y-%m-%dT%H:%M:%S')
LDFLAGS=-ldflags "-X main.buildVersion=$(VERSION) -X main.buildDate=$(DATE)"
MIGRATIONS_DIR=./migrations
GOOSE_BIN=goose
PG_DSN=${DATABASE_DSN}

# Client command variables (override on command line, e.g. make register USERNAME=bob)
CLIENT_ADDR ?= localhost:3200
USERNAME    ?= user
LOGIN       ?=
PASS        ?=
TEXT        ?=
NUMBER      ?=
EXPIRY      ?=
CVC         ?=
HOLDER      ?=
META        ?=
RECORD_ID   ?=

.PHONY: all build-server build-client test test-coverage proto lint clean run-server \
        register login \
        add-login add-text add-card add-binary \
        list get delete sync

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

# ─── Client commands ─────────────────────────────────────────────────────────
# Usage examples:
#   make register USERNAME=myuser
#   make login    USERNAME=myuser
#   make add-login LOGIN=github.com PASS=secret META="My GitHub"
#   make add-text  TEXT="some note" META=optional
#   make add-card  NUMBER=4111111111111111 EXPIRY=12/27 CVC=123 HOLDER="John Doe"
#   make list
#   make get    RECORD_ID=<uuid>
#   make delete RECORD_ID=<uuid>
#   make sync

register: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) register -u $(USERNAME)

login: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) login -u $(USERNAME)

add-login: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) add login --login "$(LOGIN)" --pass "$(PASS)" --meta "$(META)"

add-text: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) add text --text "$(TEXT)" --meta "$(META)"

add-card: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) add card --number "$(NUMBER)" --expiry "$(EXPIRY)" --cvc "$(CVC)" --holder "$(HOLDER)" --meta "$(META)"

add-binary: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) add binary --file "$(FILE)" --meta "$(META)"

list: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) list

get: build-client
	@[ "$(RECORD_ID)" ] || { echo "Usage: make get RECORD_ID=<uuid>"; exit 1; }
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) get $(RECORD_ID)

delete: build-client
	@[ "$(RECORD_ID)" ] || { echo "Usage: make delete RECORD_ID=<uuid>"; exit 1; }
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) delete $(RECORD_ID)

sync: build-client
	./$(BUILD_DIR)/$(BINARY_CLIENT) -s $(CLIENT_ADDR) sync

# Help command
help:
	@echo "Build:"
	@echo "  make all            - Build server and client"
	@echo "  make build-server   - Build server binary"
	@echo "  make build-client   - Build client binary"
	@echo "  make clean          - Remove build artifacts"
	@echo ""
	@echo "Server:"
	@echo "  make run-server     - Build and start the gRPC server (default ports)"
	@echo ""
	@echo "Client (set CLIENT_ADDR=host:port to override, default localhost:3200):"
	@echo "  make register USERNAME=<u>                          - Register new user"
	@echo "  make login    USERNAME=<u>                          - Login (refresh token)"
	@echo "  make add-login LOGIN=<l> PASS=<p> [META=<m>]       - Add login/password pair"
	@echo "  make add-text  TEXT=<t>  [META=<m>]                - Add text note"
	@echo "  make add-card  NUMBER=<n> EXPIRY=<e> CVC=<c> HOLDER=<h> [META=<m>] - Add bank card"
	@echo "  make add-binary FILE=<path> [META=<m>]             - Add binary file"
	@echo "  make list                                           - List all local records"
	@echo "  make get    RECORD_ID=<uuid>                        - Decrypt and print a record"
	@echo "  make delete RECORD_ID=<uuid>                        - Mark record as deleted"
	@echo "  make sync                                           - Sync with server"
	@echo ""
	@echo "Dev:"
	@echo "  make test           - Run all unit tests"
	@echo "  make test-coverage  - Run tests and show coverage"
	@echo "  make proto          - Generate gRPC code from .proto"
	@echo "  make lint           - Run golangci-lint"
	@echo ""
	@echo "Migrations (requires DATABASE_DSN env var):"
	@echo "  make migrate-up     - Apply all pending migrations"
	@echo "  make migrate-down   - Rollback last migration"
	@echo "  make migrate-status - Show migration status"

migrate-new:
	@echo "Creating new migration: $(name)"
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) create $(name) sql
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) fix

migrate-up:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(PG_DSN)" up

migrate-down:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(PG_DSN)" down

migrate-reset:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(PG_DSN)" reset

migrate-fix:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) fix

migrate-status:
	$(GOOSE_BIN) -dir $(MIGRATIONS_DIR) postgres "$(PG_DSN)" status
