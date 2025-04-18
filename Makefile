# Project metadata
BIN_DIR := bin
CMD_DIRS := order product
GO := go

# Go build flags
BUILD_FLAGS := -ldflags "-s -w"

.PHONY: all build run test fmt lint clean generate migrate migrate-order migrate-product

## Default target: build all binaries
all: build

## Build all binaries
build:
	@echo ">> Building binaries..."
	@mkdir -p $(BIN_DIR)
	@for dir in $(CMD_DIRS); do \
		echo ">> Building $$dir..."; \
		$(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/$$dir ./cmd/$$dir; \
	done

## Run a specific binary
run:
	@echo "Usage: make run CMD=<command> (e.g., CMD=order)"
	@false

run-%:
	@echo ">> Running $*..."
	$(GO) run ./cmd/$*

## Run tests
test:
	@echo ">> Running tests..."
	$(GO) test -v ./...

## Format code
fmt:
	@echo ">> Formatting..."
	$(GO) fmt ./...

## Lint code
lint:
	@echo ">> Linting..."
	@if ! [ -x "$$(command -v golint)" ]; then \
		echo ">> Installing golint..."; \
		$(GO) install golang.org/x/lint/golint@latest; \
	fi
	@golint ./...

## Generate code (e.g., Protobufs)
generate:
	@echo ">> Generating code..."
	buf generate

## Build migration tool
build-migrate:
	@echo ">> Building migration tool..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/migrate ./cmd/migrate

## Run migrations for all services
migrate: build-migrate
	@echo ">> Running migrations..."
	@echo "Usage: make migrate-order or make migrate-product"

## Run migrations for order service
migrate-order: build-migrate
	@echo ">> Running migrations for order service..."
	$(BIN_DIR)/migrate -service=order

## Run migrations for product service
migrate-product: build-migrate
	@echo ">> Running migrations for product service..."
	$(BIN_DIR)/migrate -service=product

## Clean binaries
clean:
	@echo ">> Cleaning up..."
	rm -rf $(BIN_DIR)
