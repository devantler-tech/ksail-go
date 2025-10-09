#!/bin/bash

# Script to generate JSON schema for KSail configuration
# This script is used by the pre-commit hook

set -e

echo "Generating JSON schema for KSail configuration..."

# Run the schema generator
pushd "$(git rev-parse --show-toplevel)" > /dev/null || exit 1

# Set GOPROXY to direct to avoid proxy issues in CI environments
export GOPROXY=direct

# Ensure dependencies are downloaded
go mod download

# Run the schema generator
go run ./cmd/schema-gen

popd > /dev/null
echo "JSON schema generation completed successfully"
