.PHONY: help test lint build ci clean install use-dev use-brew

# Determine local bin directory
LOCAL_BIN_DIR := $(or $(XDG_BIN_HOME),$(if $(wildcard $(HOME)/.local/bin),$(HOME)/.local/bin,$(HOME)/bin))
LOCAL_BIN := $(LOCAL_BIN_DIR)/git-folder

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

build: ## Build binary
	go build -v ./cmd/git-folder

ci: test lint build ## Run all CI checks locally
	@echo ""
	@echo "✓ All CI checks passed"

clean: ## Remove built artifacts
	rm -f git-folder coverage.txt

install: ## Install to $$GOPATH/bin
	go install ./cmd/git-folder

use-dev: ## Switch to local development version
	@if brew list --formula git-folder &>/dev/null && brew ls --verbose git-folder 2>/dev/null | grep -q "bin/git-folder"; then \
		echo "Unlinking brew version..."; \
		brew unlink git-folder; \
	fi
	@echo "Building and installing to $(LOCAL_BIN)..."
	@mkdir -p $(LOCAL_BIN_DIR)
	@go build -o $(LOCAL_BIN) ./cmd/git-folder
	@echo "✓ Using local version at $(LOCAL_BIN)"
	@echo "  Version: $$($(LOCAL_BIN) version)"

use-brew: ## Switch to Homebrew version
	@if [ ! -f "$(LOCAL_BIN)" ]; then \
		echo "No local version installed"; \
	else \
		echo "Removing local version..."; \
		rm $(LOCAL_BIN); \
		echo "✓ Removed local version"; \
	fi
	@if ! brew list --formula git-folder &>/dev/null; then \
		echo "Error: git-folder not installed via brew"; \
		echo "Install with: brew install claybridges/tap/git-folder"; \
		exit 1; \
	fi
	@echo "Linking brew version..."
	@brew link git-folder 2>/dev/null || true
	@echo "✓ Using brew version"
	@echo "  Version: $$(git-folder version)"
