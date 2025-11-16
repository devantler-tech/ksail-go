#!/bin/bash

# Script to ensure golangci-lint is installed and run it with --fix
# This script is used by the pre-commit hook

set -e

# Function to check if golangci-lint is available and config is compatible
check_golangci_lint() {
	# Save current directory
	local original_dir
	original_dir="$(pwd)"

	# Change to src directory where go.mod is located for config check (if it exists)
	if [ -d "src" ]; then
		cd src
	fi

	local result=0
	if command -v golangci-lint >/dev/null 2>&1; then
		# Test if golangci-lint config is compatible with current project
		if golangci-lint config path >/dev/null 2>&1; then
			result=0
		else
			result=2 # golangci-lint exists but config is incompatible
		fi
	elif [ -x "$HOME/go/bin/golangci-lint" ]; then
		# Test if golangci-lint config is compatible with current project
		if "$HOME/go/bin/golangci-lint" config path >/dev/null 2>&1; then
			result=0
		else
			result=2 # golangci-lint exists but config is incompatible
		fi
	else
		result=1 # golangci-lint not found
	fi

	# Return to original directory
	cd "$original_dir"
	return $result
}

# Function to attempt golangci-lint installation
install_golangci_lint() {
	echo "golangci-lint not found. Attempting to install..."
	echo ""
	echo "Note: Using official binary installation method (recommended by golangci-lint)."
	echo "See: https://golangci-lint.run/welcome/install/"
	echo ""

	# Use official binary installation script
	echo "Installing golangci-lint via official binary install script..."
	if curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$HOME/go/bin" latest; then
		echo "Installation successful."
	else
		echo "Automatic installation failed."
		echo ""
		echo "Please install golangci-lint manually from:"
		echo "https://golangci-lint.run/usage/install/"
		echo ""
		return 1
	fi
	echo ""
}

# Function to run golangci-lint
run_golangci_lint() {
	echo "Running golangci-lint run --fix..."

	# Save current directory and change to src directory where go.mod is located (if it exists)
	local original_dir
	original_dir="$(pwd)"
	if [ -d "src" ]; then
		cd src
	fi

	local exit_code=0
	if command -v golangci-lint >/dev/null 2>&1; then
		golangci-lint run --fix || exit_code=$?
	elif [ -x "$HOME/go/bin/golangci-lint" ]; then
		"$HOME/go/bin/golangci-lint" run --fix || exit_code=$?
	else
		echo "Error: golangci-lint not found after installation"
		cd "$original_dir"
		exit 1
	fi

	# Return to original directory
	cd "$original_dir"

	if [ $exit_code -ne 0 ]; then
		exit $exit_code
	fi

	echo "golangci-lint completed successfully"
}

# Main execution
main() {
	if ! check_golangci_lint; then
		check_result=$?
		case $check_result in
		1) # golangci-lint not found
			if ! install_golangci_lint; then
				exit 1
			fi
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
			;;
		2) # golangci-lint exists but incompatible
			echo "golangci-lint is installed but appears to be incompatible with the project configuration."
			echo ""
			echo "Please reinstall golangci-lint from:"
			echo "https://golangci-lint.run/usage/install/"
			echo ""
			echo "Then run: golangci-lint run --fix"
			exit 1
			;;
		esac
	else
		# golangci-lint is available and compatible
		run_golangci_lint
	fi
}

# Run main function
main "$@"
