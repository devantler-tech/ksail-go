#!/bin/bash

# Script to generate JSON schema for KSail configuration
# This script is used by the pre-commit hook

set -e

echo "Generating JSON schema for KSail configuration..."

# Run the schema generator
pushd "$(git rev-parse --show-toplevel)" > /dev/null || exit 1

# Set GOPROXY to direct to avoid proxy issues in CI environments
export GOPROXY=direct

# Use local Go toolchain instead of trying to download a specific version
# This is important for CI environments that may not have the exact Go version
export GOTOOLCHAIN=local

# Check Go version and handle tool directive compatibility
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

# If Go version is less than 1.24, temporarily remove tool directive
if [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 24 ]; then
    echo "Detected Go version < 1.24, temporarily handling tool directive..."
    # Create backup
    cp go.mod go.mod.bak
    # Remove tool directive line
    grep -v "^tool " go.mod.bak > go.mod || true
fi

# Ensure dependencies are downloaded
go mod download

# Run the schema generator
go run ./cmd/schema-gen

# Restore original go.mod if we modified it
if [ -f go.mod.bak ]; then
    mv go.mod.bak go.mod
fi

popd > /dev/null
echo "JSON schema generation completed successfully"
