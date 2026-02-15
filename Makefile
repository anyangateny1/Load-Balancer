APP_NAME := lb
CMD_PATH := ./cmd/lb
BUILD_DIR := bin

GO := go
LINT := golangci-lint

.PHONY: all build run test lint fmt vet clean tidy

all: fmt vet lint test build

build:
	@echo "Building..."
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(CMD_PATH)

run:
	$(GO) run $(CMD_PATH)

test:
	@echo "Running tests..."
	$(GO) test ./... -v -race

lint:
	@echo "Running linter..."
	$(LINT) run

fmt:
	@echo "Formatting..."
	goimports -w .
	$(GO) fmt ./...

vet:
	@echo "Running go vet..."
	$(GO) vet ./...

tidy:
	@echo "Tidying modules..."
	$(GO) mod tidy

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
