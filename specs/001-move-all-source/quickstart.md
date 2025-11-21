# Quickstart Guide: Move All Go Source Code to src/

**Feature**: Move All Go Source Code to src/
**Date**: 2025-11-16
**Audience**: Developers implementing this reorganization

> **⚠️ NOTE**: This feature was **NEVER IMPLEMENTED**. This quickstart guide describes how the implementation would be performed, but the actual code reorganization did not happen. The codebase still has all Go source code at the repository root. If you wish to implement this reorganization, follow the steps below.

## Prerequisites

- Go 1.25.4+ installed
- Git installed with `git mv` support
- Write access to the repository
- Clean git working directory (no uncommitted changes)

## Implementation Steps

### Phase 1: Pre-Move Validation

Run baseline validation to establish reference metrics:

```bash
# Run from repository root
cd /path/to/ksail-go

# Verify clean state
git status

# Run baseline checks
go build ./...
go test ./...
golangci-lint run --timeout 5m

# Capture baseline build time
time go build -o bin/ksail .
```

**Expected**: All commands succeed with exit code 0. Note the build time for comparison.

### Phase 2: File Reorganization

Create src/ directory and move files:

```bash
# Create src directory
mkdir src

# Move Go source directories
git mv cmd src/
git mv pkg src/

# Move main files
git mv main.go src/
git mv main_test.go src/

# Move Go module files
git mv go.mod src/
git mv go.sum src/
```

**Expected**: All files moved with git tracking file renames.

### Phase 3: Configuration Updates

Update configuration files to reference new paths:

#### VS Code Tasks (`.vscode/tasks.json`)

```json
{
  "tasks": [
    {
      "label": "go: build",
      "command": "go",
      "args": ["build", "./..."],
      "options": {
        "cwd": "${workspaceFolder}/src"  // Changed from ${workspaceFolder}
      }
    },
    {
      "label": "go: test",
      "command": "go",
      "args": ["test", "./..."],
      "options": {
        "cwd": "${workspaceFolder}/src"  // Changed from ${workspaceFolder}
      }
    }
  ]
}
```

#### GitHub Workflows (`.github/workflows/ci.yaml`)

```yaml
jobs:
  build:
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: 'src/go.mod'  # Changed from 'go.mod'
      - name: Build
        working-directory: src  # Added
        run: go build ./...
      - name: Test
        working-directory: src  # Added
        run: go test ./...
```

#### GoReleaser (`.goreleaser.yaml`)

```yaml
builds:
  - id: ksail
    main: ./src  # Changed from '.'
    binary: ksail
    env:
      - CGO_ENABLED=0
```

#### Mockery (`.mockery.yml`)

Update paths if they reference absolute locations:

```yaml
# If mockery config references source paths, update them to include src/ prefix
```

#### Script Updates

For any scripts in `.github/scripts/` that reference Go source:

```bash
# Example: generate-schema.sh
cd src  # Add this at the start
go run . gen schema
```

### Phase 4: Post-Move Validation

Verify everything works with new structure:

```bash
# Run from repository root
cd src

# Build from new location
go build ./...

# Test from new location
go test ./...

# Return to root for linting
cd ..
golangci-lint run --timeout 5m

# Build binary to correct output location
cd src
go build -o ../bin/ksail .

# Verify binary exists
ls -l ../bin/ksail

# Test mockery
cd ..
mockery
```

**Expected**: All commands succeed, build time within 5% of baseline.

### Phase 5: Commit Changes

```bash
# Stage all changes
git add -A

# Commit with clear message
git commit -m "feat: reorganize source code into src/ directory

Moves all Go source code (cmd/, pkg/, internal/, main.go, main_test.go)
and Go module files (go.mod, go.sum) into a new src/ subdirectory.

Updates configuration files to reference new paths:
- .vscode/tasks.json: Updated working directory
- .github/workflows/*.yaml: Added working-directory and updated go.mod path
- .goreleaser.yaml: Updated main path
- Scripts: Added cd src where needed

This structural change maintains 100% backward compatibility for external
consumers. Import paths remain unchanged: github.com/devantler-tech/ksail-go/pkg/...

All validation checks passed:
- Build: ✓
- Tests: ✓
- Lint: ✓
- Build time: <baseline> (unchanged)

Resolves #<issue-number>"
```

### Phase 6: Pre-Merge Validation

Push to feature branch and verify CI:

```bash
# Push to feature branch
git push origin 001-move-all-source

# Monitor GitHub Actions
# Verify all workflows pass:
# - CI workflow (build, test, lint)
# - Any other automated checks
```

**Manual Checks**:

1. Open project in VS Code and reload workspace
2. Verify autocomplete and navigation work
3. Run VS Code tasks: build, test, fmt, lint
4. Test external import (optional):

   ```bash
   mkdir /tmp/test-import
   cd /tmp/test-import
   go mod init test
   go get github.com/devantler-tech/ksail-go@<your-branch>
   # Create test file importing a package
   go build
   ```

### Phase 7: Merge and Post-Merge Validation

After PR approval:

```bash
# Merge PR
# Via GitHub UI or:
git checkout main
git pull origin main

# Verify post-merge
cd src
go build ./...
go test ./...

# Monitor CI/CD
# Check GitHub Actions for main branch
# Verify all workflows pass
```

## Quick Reference Commands

**Build**:

```bash
cd src && go build ./...
# OR
go -C src build ./...
```

**Test**:

```bash
cd src && go test ./...
# OR
go -C src test ./...
```

**Run**:

```bash
cd src && go run .
# OR
go -C src run .
```

**Build binary**:

```bash
cd src && go build -o ../bin/ksail .
```

## Rollback Procedure

If validation fails at any step:

```bash
# If pre-commit (not yet committed)
git reset --hard HEAD
git clean -fd

# If post-commit but not merged
git reset --hard HEAD~1

# If post-merge
git revert -m 1 <merge-commit-sha>
git push origin main
```

Then document the failure and create a new PR with fixes.

## Success Indicators

✅ All Go commands work from src/ directory
✅ VS Code workspace recognizes module after reload
✅ CI/CD pipelines pass on feature branch
✅ Build times unchanged
✅ Git history preserved (`git log --follow src/main.go` shows full history)
✅ External imports work without changes

## Notes for Developers

After this reorganization:

- **Local development**: Work in src/ directory or use `-C src` flag
- **IDE**: Reload workspace after pulling these changes
- **New files**: Create under src/ (src/cmd/, src/pkg/, etc.)
- **Imports**: Unchanged - still `github.com/devantler-tech/ksail-go/pkg/...`
- **Build output**: Binaries still go to bin/ at repository root
