#!/bin/bash

# Script to create refactor package issues for all packages in pkg/
# This script requires GitHub CLI (gh) to be installed and authenticated

set -e

REPO="devantler-tech/ksail-go"

# Check if gh is installed and authenticated
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

if ! gh auth status &> /dev/null; then
    echo "Error: GitHub CLI is not authenticated"
    echo "Run: gh auth login"
    exit 1
fi

# Array of packages to create issues for
packages=("apis" "client" "cmd" "di" "io" "k8s" "svc" "testutils" "ui" "workload")

# Issue body template
create_issue_body() {
    local package_name=$1
    cat <<EOF
### Description

Refactor the \`pkg/${package_name}\` package to improve code quality, maintainability, and adherence to project standards.

### Tasks

- [ ] Review code for adherence to SOLID principles
- [ ] Eliminate code duplication (target: 0% duplication per jscpd config)
- [ ] Improve error handling and propagation
- [ ] Add or update documentation comments for exported types and functions
- [ ] Review and optimize package structure
- [ ] Ensure proper test coverage
- [ ] Fix any linting issues
- [ ] Address code smells (bloaters, couplers, etc.)

### Acceptance Criteria

- All code passes linting (golangci-lint, mega-linter)
- Zero code duplication as measured by jscpd
- All exported types and functions have proper documentation
- Test coverage maintained or improved
- Code follows repository conventions and style guidelines
EOF
}

echo "Creating refactor issues for packages in pkg/..."
echo "Repository: $REPO"
echo ""

for package in "${packages[@]}"; do
    title="chore: refactor pkg/${package} package"
    body=$(create_issue_body "$package")
    labels="chore,refactor,package:${package}"
    
    echo "Creating issue for pkg/${package}..."
    
    # Create the issue
    if gh issue create \
        --repo "$REPO" \
        --title "$title" \
        --body "$body" \
        --label "$labels"; then
        echo "✓ Successfully created issue for pkg/${package}"
    else
        echo "✗ Failed to create issue for pkg/${package}"
    fi
    
    echo ""
    
    # Small delay to avoid rate limiting
    sleep 1
done

echo "Done! Created issues for ${#packages[@]} packages."
