# Pre-commit Configuration

This repository uses the [pre-commit framework](https://pre-commit.com/) for automatic code formatting and quality checks before each commit.

## Features

The pre-commit configuration automatically:

- **Removes trailing whitespace** from Go files
- **Ensures files end with newlines** (Go standard requirement)
- **Runs golangci-lint with auto-fix** to format and lint Go code according to project standards

## Installation

The pre-commit hooks are automatically installed when you clone this repository. If you need to reinstall them:

```bash
# Install pre-commit (if not already installed)
pip install pre-commit

# Install the git hook scripts
pre-commit install
```

## Requirements

- **Python 3.6+** and **pip** for the pre-commit framework
- **golangci-lint** must be installed and available in your PATH or at `~/go/bin/golangci-lint`

### Installing golangci-lint

If golangci-lint is not installed:

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

## Usage

### Automatic Execution

Pre-commit hooks run automatically on every `git commit`. If issues are found:

1. **Formatting fixes are applied automatically** (trailing whitespace, missing newlines)
2. **golangci-lint fixes are applied** (imports, formatting, some linting issues)
3. **The commit is blocked** if there are unfixable issues
4. **You must review and stage the changes**, then commit again

### Manual Execution

You can run pre-commit checks manually:

```bash
# Run on all files
pre-commit run --all-files

# Run on specific files
pre-commit run --files file1.go file2.go

# Run a specific hook
pre-commit run golangci-lint --all-files
```

### Example Workflow

```bash
# Make your changes
git add .

# Attempt to commit
git commit -m "Your commit message"

# If pre-commit fixes issues, review the changes
git diff

# Stage the fixes and commit again
git add .
git commit -m "Your commit message"
```

## Configuration

The pre-commit configuration is in `.pre-commit-config.yaml`. It uses local hooks to ensure compatibility with your installed golangci-lint version and project configuration.

### Hooks Included

1. **trailing-whitespace**: Removes trailing whitespace from Go files
2. **end-of-file-fixer**: Ensures Go files end with a newline
3. **golangci-lint**: Runs `golangci-lint run --fix` using your local installation

## Bypassing Hooks

In rare cases where you need to bypass the hooks (not recommended):

```bash
git commit --no-verify -m "Your message"
```

**Warning**: Bypassing hooks may introduce formatting inconsistencies and should only be used in exceptional circumstances.

## Troubleshooting

### Pre-commit not found

If you see "pre-commit: command not found":

```bash
# Install pre-commit
pip install pre-commit

# Reinstall hooks
pre-commit install
```

### golangci-lint not found

If the hook fails with "golangci-lint not found":

1. Install golangci-lint as shown above
2. Ensure it's in your PATH or at `~/go/bin/golangci-lint`
3. Test manually: `golangci-lint version`

### Hook installation fails

If hook installation fails, try:

```bash
# Clean pre-commit cache
pre-commit clean

# Reinstall
pre-commit install
```

## Benefits

- **Consistent formatting** across all contributors
- **Automatic fixes** for common issues
- **Prevents bad commits** from entering the repository
- **Reduces code review time** by catching formatting issues early
- **Uses your local tooling** ensuring consistency with your development environment