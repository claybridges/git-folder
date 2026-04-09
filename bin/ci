#!/bin/bash
# Run all CI checks locally

set -e

echo "==> Running tests..."
go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

echo ""
echo "==> Running linter..."
golangci-lint run ./...

echo ""
echo "==> Building binary..."
go build -v ./cmd/git-folder

echo ""
echo "==> Testing binary..."
./git-folder version

echo ""
echo "✓ All CI checks passed"
