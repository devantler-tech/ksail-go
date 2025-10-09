#!/bin/bash

# Script to generate JSON schema for KSail configuration
# This script is used by the pre-commit hook
# Note: Requires Go 1.24+ due to the 'tool' directive in go.mod

set -e

echo "Generating JSON schema for KSail configuration..."

# Run the schema generator
pushd "$(git rev-parse --show-toplevel)" > /dev/null || exit 1

# Run the schema generator
go run ./cmd/schema-gen

popd > /dev/null
echo "JSON schema generation completed successfully"
