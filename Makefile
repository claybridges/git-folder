.PHONY: help test lint build ci clean install

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
