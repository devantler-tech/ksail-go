#!/bin/bash

# Script to ensure golangci-lint is available and run it
# This script is used by the pre-commit hook

set -e

# Function to check if golangci-lint is available
check_golangci_lint() {
    if command -v golangci-lint >/dev/null 2>&1; then
        return 0
    elif [ -x "$HOME/go/bin/golangci-lint" ]; then
        return 0
    else
        return 1  # golangci-lint not found
    fi
}

# Function to run golangci-lint
run_golangci_lint() {
    if command -v golangci-lint >/dev/null 2>&1; then
        golangci-lint run --new-from-rev HEAD --fix
    elif [ -x "$HOME/go/bin/golangci-lint" ]; then
        "$HOME/go/bin/golangci-lint" run --new-from-rev HEAD --fix
    else
        echo "Error: golangci-lint not found"
        echo "Please install golangci-lint. See:"
        echo "https://golangci-lint.run/usage/install/"
        exit 1
    fi
}

# Main execution
main() {
    if ! check_golangci_lint; then
        echo "golangci-lint not found. Please install it first."
        echo "Installation instructions: https://golangci-lint.run/usage/install/"
        exit 1
    else
        # golangci-lint is available
        run_golangci_lint
    fi
}

# Run main function
main "$@"