#!/bin/bash

# Script to generate JSON schema for KSail configuration
# This script is used by the pre-commit hook

set -e

echo "Generating JSON schema for KSail configuration..."

# Run the schema generator
cd "$(git rev-parse --show-toplevel)" || exit 1
go run ./cmd/schema-gen

echo "JSON schema generation completed successfully"
