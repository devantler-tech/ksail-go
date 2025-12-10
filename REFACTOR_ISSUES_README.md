# Refactor Package Issues

This directory contains resources for creating refactor issues for each package in `pkg/`.

## Files

- **refactor-package-issues.md** - Complete content for all 10 refactor issues, one per package
- **create-refactor-issues.sh** - Automated script to create all issues via GitHub CLI
- **README.md** - This file

## Usage

### Option 1: Automated Creation (Recommended)

If you have GitHub CLI installed and authenticated:

```bash
./create-refactor-issues.sh
```

This will create all 10 issues automatically with proper titles, labels, and content.

### Option 2: Manual Creation

If you prefer to create issues manually or need to review them first:

1. Open `refactor-package-issues.md`
2. For each issue section:
   - Go to https://github.com/devantler-tech/ksail-go/issues/new
   - Select the "chore" template (if available)
   - Copy the title, body content, and labels
   - Submit the issue

### Option 3: GitHub CLI (Individual Issues)

To create issues one at a time:

```bash
gh issue create \
  --repo devantler-tech/ksail-go \
  --title "chore: refactor pkg/<package-name> package" \
  --label "chore,refactor,package:<package-name>" \
  --body "$(cat << 'EOF'
<paste body content from refactor-package-issues.md>
EOF
)"
```

## Packages Covered

The following 10 packages have issue templates ready:

1. pkg/apis
2. pkg/client
3. pkg/cmd
4. pkg/di
5. pkg/io
6. pkg/k8s
7. pkg/svc
8. pkg/testutils
9. pkg/ui
10. pkg/workload

## Issue Template Structure

Each issue follows the chore template format with:

- **Title**: Semantic commit format (`chore: refactor pkg/<name> package`)
- **Labels**: `chore`, `refactor`, `package:<name>`
- **Description**: Overview of refactoring goals
- **Tasks**: Checklist covering:
  - SOLID principles adherence
  - Code duplication elimination (0% target)
  - Error handling improvements
  - Documentation updates
  - Package structure optimization
  - Test coverage
  - Linting fixes
  - Code smell remediation
- **Acceptance Criteria**: Clear success metrics

## Requirements

For automated creation:

- [GitHub CLI](https://cli.github.com/) installed
- GitHub CLI authenticated: `gh auth login`
- Write access to devantler-tech/ksail-go repository

## Notes

- All issues target the same quality standards defined in the repository's coding guidelines
- The 0% code duplication target aligns with the `.jscpd.json` configuration
- Each issue is scoped to a single package for focused refactoring efforts
