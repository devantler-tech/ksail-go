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

# Function to attempt golangci-lint installation
install_golangci_lint() {
    echo "golangci-lint not found. Attempting to install..."
    echo ""
    echo "Installing golangci-lint v2.4.0 via curl..."
    
    # Install golangci-lint v2.4.0 using the official installation script
    # This ensures we get a compatible v2.x version
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$HOME/go/bin" v2.4.0
    
    echo ""
    echo "If installation fails, please install golangci-lint manually:"
    echo "https://golangci-lint.run/usage/install/"
    echo ""
}

# Function to run golangci-lint
run_golangci_lint() {
    echo "Running golangci-lint..."
    
    if command -v golangci-lint >/dev/null 2>&1; then
        golangci-lint run --new-from-rev HEAD --fix
    elif [ -x "$HOME/go/bin/golangci-lint" ]; then
        "$HOME/go/bin/golangci-lint" run --new-from-rev HEAD --fix
    else
        echo "Error: golangci-lint not found after installation"
        echo "Please install golangci-lint manually:"
        echo "https://golangci-lint.run/usage/install/"
        exit 1
    fi
    
    echo "golangci-lint completed successfully"
}

# Main execution
main() {
    if ! check_golangci_lint; then
        install_golangci_lint
        echo "Attempting to run golangci-lint after installation..."
        if ! run_golangci_lint; then
            echo ""
            echo "Installation completed but golangci-lint failed to run properly."
            echo "Please install golangci-lint manually from:"
            echo "https://golangci-lint.run/usage/install/"
            exit 1
        fi
    else
        # golangci-lint is available
        run_golangci_lint
    fi
}

# Run main function
main "$@"