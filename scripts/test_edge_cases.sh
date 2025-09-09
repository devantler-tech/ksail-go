#!/bin/bash
# test_edge_cases.sh - Script to achieve 100% code coverage for asciiart package
# This script demonstrates how to test edge cases in internal functions through
# the public API by temporarily modifying the embedded logo file and rebuilding.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LOGO_FILE="$PROJECT_ROOT/cmd/ui/asciiart/ksail_logo.txt"
BACKUP_FILE="$LOGO_FILE.backup"

echo "=== Testing Edge Cases for ASCII Art Coverage ==="

# Function to restore original file
restore_logo() {
    if [ -f "$BACKUP_FILE" ]; then
        echo "Restoring original logo file..."
        mv "$BACKUP_FILE" "$LOGO_FILE"
    fi
}

# Set up cleanup trap
trap restore_logo EXIT

# Backup original logo
echo "Backing up original logo file..."
cp "$LOGO_FILE" "$BACKUP_FILE"

# Create logo content that triggers edge cases
echo "Creating edge case logo content..."
cat > "$LOGO_FILE" << 'EOF'
                    __ ______     _ __
                   / //_/ __/__ _(_) /
                  / ,< _\ \/ _ `/ / /
                 /_/|_/___/\_,_/_/_/
                                   . . .
short_line_edge_case
        _____/______|             ___|____     |"\/"|
short_cyan
\   -----       -\-\-\-    |    |  ^        \___/  |
~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~^~
EOF

echo "Edge case logo created with:"
echo "- Line 6: $(sed -n '6p' "$LOGO_FILE" | wc -c) chars (< 38, triggers printGreenBlueCyanPart edge case)"
echo "- Line 8: $(sed -n '8p' "$LOGO_FILE" | wc -c) chars (< 32, triggers printGreenCyanPart edge case)"

# Build and test with edge case logo
echo ""
echo "Building and testing with edge case logo..."
cd "$PROJECT_ROOT"

# Run tests with coverage
go test -v -coverprofile=coverage_edge.out ./cmd/ui/asciiart

# Check coverage
echo ""
echo "=== Coverage Report ==="
go tool cover -func=coverage_edge.out | grep ksail_logo

echo ""
echo "=== Total Coverage ==="
go tool cover -func=coverage_edge.out | tail -1

# Clean up coverage file
rm -f coverage_edge.out

echo ""
echo "Edge case testing completed!"
echo "The original logo file will be restored automatically."