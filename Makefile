# Makefile
.PHONY: build clean test install dev help

# Variables
APP_NAME := execute-my-will
VERSION := $(shell cat VERSION 2>/dev/null || echo "0.1.0")
BUILD_DIR := dist
CMD_DIR := cmd/$(APP_NAME)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

# Build targets
PLATFORMS := linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)$(APP_NAME) - Build System$(NC)"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf $(BUILD_DIR)
	go clean -cache

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v -race -coverprofile=coverage.out ./test/
	go tool cover -html=coverage.out -o coverage.html

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...

deps: ## Download dependencies
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	go mod download
	go mod tidy

build: clean ## Build for current platform
	@echo "$(BLUE)Building $(APP_NAME) v$(VERSION) for current platform...$(NC)"
	mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(APP_NAME)$(NC)"

build-all: clean ## Build for all platforms
	@echo "$(BLUE)Building $(APP_NAME) v$(VERSION) for all platforms...$(NC)"
	mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		output="$(BUILD_DIR)/$(APP_NAME)-$(VERSION)-$$os-$$arch$$ext"; \
		echo "$(YELLOW)Building for $$os/$$arch...$(NC)"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build $(LDFLAGS) -o $$output ./$(CMD_DIR); \
		if [ $$? -eq 0 ]; then \
			echo "$(GREEN)✓ Built: $$output$(NC)"; \
			cd $(BUILD_DIR) && sha256sum $$(basename $$output) > $$(basename $$output).sha256 && cd ..; \
		else \
			echo "$(RED)✗ Failed to build for $$os/$$arch$(NC)"; \
		fi; \
	done
	@echo "$(GREEN)All builds complete!$(NC)"

install: build ## Install binary to $GOPATH/bin
	@echo "$(BLUE)Installing $(APP_NAME)...$(NC)"
	go install $(LDFLAGS) ./$(CMD_DIR)
	@echo "$(GREEN)Installed $(APP_NAME) to $(shell go env GOPATH)/bin$(NC)"

dev: ## Run in development mode
	@echo "$(BLUE)Running in development mode...$(NC)"
	go run ./$(CMD_DIR) $(ARGS)



check: test vet ## Run all checks (test, vet)

ci: deps check build-all ## Run CI pipeline locally

release-check: ## Check if ready for release
	@echo "$(BLUE)Checking release readiness...$(NC)"
	@if [ ! -f VERSION ]; then echo "$(RED)VERSION file not found$(NC)"; exit 1; fi
	@echo "$(GREEN)Version: $(VERSION)$(NC)"
	@echo "$(YELLOW)Checking git status...$(NC)"
	@git diff --quiet || (echo "$(RED)Working directory is dirty$(NC)" && exit 1)
	@echo "$(GREEN)Ready for release!$(NC)"

update-version: ## Update version (usage: make update-version VERSION=1.2.3)
	@if [ -z "$(VERSION)" ]; then echo "$(RED)Please specify VERSION$(NC)"; exit 1; fi
	@echo "$(BLUE)Updating