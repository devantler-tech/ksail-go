#!/bin/bash

# Script to ensure mockery is installed and run it
# This script is used by the pre-commit hook

set -e

# Function to check if mockery is available and working
check_mockery() {
    if command -v mockery >/dev/null 2>&1; then
        # Test if mockery can run with current config
        if mockery --dry-run >/dev/null 2>&1; then
            return 0
        else
            return 2  # mockery exists but config is incompatible
        fi
    elif [ -x "$HOME/go/bin/mockery" ]; then
        # Test if mockery can run with current config
        if "$HOME/go/bin/mockery" --dry-run >/dev/null 2>&1; then
            return 0
        else
            return 2  # mockery exists but config is incompatible
        fi
    else
        return 1  # mockery not found
    fi
}

# Function to attempt mockery installation
install_mockery() {
    echo "Mockery not found. Attempting to install..."
    echo ""
    echo "Note: This project requires mockery v3.x for compatibility with the configuration."
    echo "The script will attempt automatic installation, but manual installation may be required."
    echo ""
    
    # Try go install for mockery v3.x
    echo "Installing mockery v3.x via go install..."
    go install github.com/vektra/mockery/v3@latest
    
    echo ""
    echo "If you encounter configuration errors, please see the manual installation instructions for mockery v3.x:"
    echo "https://vektra.github.io/mockery/v3.5/installation/"
    echo ""
}

# Function to run mockery
run_mockery() {
    echo "Running mockery to generate mocks..."
    
    if command -v mockery >/dev/null 2>&1; then
        mockery
    elif [ -x "$HOME/go/bin/mockery" ]; then
        "$HOME/go/bin/mockery"
    else
        echo "Error: mockery not found after installation"
        exit 1
    fi
    
    echo "Mockery completed successfully"
}

# Main execution
main() {
    if ! check_mockery; then
        check_result=$?
        case $check_result in
            1)  # mockery not found
                install_mockery
                echo "Attempting to run mockery after installation..."
                if ! run_mockery; then
                    echo ""
                    echo "Installation completed but mockery failed to run properly."
                    echo "Please install mockery v3.x manually from:"
                    echo "https://vektra.github.io/mockery/v3.5/installation/"
                    echo ""
                    echo "Then run: mockery"
                    exit 1
                fi
                ;;
            2)  # mockery exists but incompatible
                echo "Mockery is installed but appears to be incompatible with the project configuration."
                echo "This project requires mockery v3.x."
                echo ""
                echo "Please install mockery v3.x manually from:"
                echo "https://vektra.github.io/mockery/v3.5/installation/"
                echo ""
                echo "Then run: mockery"
                exit 1
                ;;
        esac
    else
        # mockery is available and compatible
        run_mockery
    fi
}

# Run main function
main "$@"