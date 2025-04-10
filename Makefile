# Project metadata
BIN_DIR := bin
CMD_DIRS := order product
GO := go

# Go build flags
BUILD_FLAGS := -ldflags "-s -w"

.PHONY: all build run test fmt lint clean generate

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

## Clean binaries
clean:
	@echo ">> Cleaning up..."
	rm -rf $(BIN_DIR)
