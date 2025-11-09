#!/bin/bash

# Script to generate JSON schema from KSail config types
# This script is used by the pre-commit hook

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SCHEMA_GENERATOR_DIR="$SCRIPT_DIR/generate-schema"

# Change to repository root
cd "$REPO_ROOT"

echo "Generating JSON schema from KSail config types..."

# Run the schema generator
cd "$SCHEMA_GENERATOR_DIR"
go run main.go "$REPO_ROOT/schemas/ksail-config.schema.json"

echo "JSON schema generation completed successfully"
