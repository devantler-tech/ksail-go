#!/bin/bash

# Script to ensure golangci-lint is installed and run it with --fix
# This script is used by the pre-commit hook

set -e

# Function to check if golangci-lint is available
check_golangci_lint() {
	if command -v golangci-lint >/dev/null 2>&1; then
		return 0
	else
		return 1
	fi
}

# Function to attempt golangci-lint installation
install_golangci_lint() {
	echo "golangci-lint not found. Attempting to install..."
	echo ""
	echo "Installing golangci-lint via go install..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	echo ""
	echo "If installation failed, please see the manual installation instructions:"
	echo "https://golangci-lint.run/usage/install/"
	echo ""
}

# Function to run golangci-lint
run_golangci_lint() {
	echo "Running golangci-lint run --fix..."

	if command -v golangci-lint >/dev/null 2>&1; then
		golangci-lint run --fix
	elif [ -x "$HOME/go/bin/golangci-lint" ]; then
		"$HOME/go/bin/golangci-lint" run --fix
	else
		echo "Error: golangci-lint not found after installation"
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
			echo ""
			echo "Then run: golangci-lint run --fix"
			exit 1
		fi
	else
		# golangci-lint is available
		run_golangci_lint
	fi
}

# Run main function
main "$@"
