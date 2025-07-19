# Project metadata
BIN_DIR := bin
CMD_DIRS := order product
GO := go

# Go build flags
BUILD_FLAGS := -ldflags "-s -w"

.PHONY: all build run-% test fmt lint clean generate docker-rebuild docker-recreate create-connectors

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
	@if ! [ -x "$$(command -v golangci-lint)" ]; then \
		echo ">> Installing golangci-lint..."; \
		$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		export PATH="$$PATH:$$($(GO) env GOPATH)/bin"; \
	fi
	@GOPATH=$$($(GO) env GOPATH) PATH="$$PATH:$$($(GO) env GOPATH)/bin" golangci-lint run

## Generate code (e.g., Protobufs)
generate:
	@echo ">> Generating code..."
	buf generate

## Clean binaries
clean:
	@echo ">> Cleaning up..."
	rm -rf $(BIN_DIR)

## Rebuild Docker images
docker-rebuild:
	@echo ">> Rebuilding Docker images..."
	docker-compose build order product

## Recreate Docker Compose environment
docker-recreate:
	@echo ">> Recreating Docker Compose environment..."
	docker-compose down
	docker-compose up -d --force-recreate

## Create and register Debezium connectors
create-connectors:
	@echo ">> Creating Debezium connectors..."
	@echo ">> Waiting for Debezium Connect to be ready..."
	@until docker-compose exec -T debezium curl -s http://localhost:8083/ > /dev/null; do \
		echo "Waiting for Debezium Connect..."; \
		sleep 5; \
	done
	@echo ">> Debezium Connect is ready."
	@echo ">> Copying connector configuration to Debezium container..."
	@docker cp config/connectors/debezium-connector-config.json go-bootiful-ordering-debezium:/app/debezium-connector-config.json
	@echo ">> Registering connector..."
	@docker-compose exec -T debezium curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" \
		http://localhost:8083/connectors/ -d @/app/debezium-connector-config.json
	@echo ">> Connector registration completed."
