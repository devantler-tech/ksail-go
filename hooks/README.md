# Git Hooks

This directory contains git hooks that are automatically linked from `.git/hooks` to ensure consistent code quality and formatting.

## Available Hooks

### pre-commit

The `pre-commit` hook runs before each commit to automatically format Go code using `golangci-lint fmt`.

**What it does:**
- Runs `golangci-lint fmt` on all Go files in the repository
- Automatically fixes formatting issues like:
  - Missing newlines at end of files
  - Extra blank lines
  - Incorrect indentation
  - Other Go formatting standards

**Behavior:**
- If no formatting changes are needed, the commit proceeds normally
- If formatting changes are made, the commit is blocked and you're prompted to review and add the changes
- Shows a list of files that were modified

**Installation:**
The hook is automatically installed via a symlink from `.git/hooks/pre-commit` to `hooks/pre-commit` when you clone the repository.

**Requirements:**
- `golangci-lint` must be installed and available in your PATH or at `~/go/bin/golangci-lint`
- If not found, the hook will display installation instructions

**Manual Installation:**
If the symlink is missing, you can recreate it:
```bash
ln -sf ../../hooks/pre-commit .git/hooks/pre-commit
```

**Bypassing the Hook:**
In rare cases where you need to bypass the hook (not recommended), you can use:
```bash
git commit --no-verify
```

## Adding New Hooks

When adding new hooks:
1. Create the hook script in this directory
2. Make it executable: `chmod +x hooks/hook-name`
3. Create a symlink from `.git/hooks/hook-name` to `hooks/hook-name`
4. Test the hook before committing
5. Update this README with documentation for the new hook