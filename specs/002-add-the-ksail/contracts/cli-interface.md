# CLI Command Contract: ksail init

## Command Interface

### Basic Usage

```bash
ksail init [flags]
```

### Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--name` | `-n` | string | current-dir | Project name for configuration |
| `--distribution` | `-d` | string | kind | Kubernetes distribution (kind, k3d, eks) |
| `--reconciliation-tool` | `-r` | string | "" | GitOps tool (kubectl, flux) |
| `--source-directory` | `-s` | string | "." | Target directory for files |
| `--force` | `-f` | boolean | false | Overwrite existing files |

| `--help` | `-h` | boolean | false | Show help information |

### Exit Codes

| Code | Meaning | Example Scenario |
|------|---------|------------------|
| 0 | Success | Project initialized successfully |
| 1 | General error | Invalid flag combination |
| 2 | File system error | Permission denied, disk full |
| 3 | Validation error | Invalid project name |
| 4 | Conflict error | Existing files without --force |

## Input Validation

### Project Name

- MUST match DNS-1123 subdomain format: `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
- MUST be 1-63 characters in length
- MUST NOT be empty or whitespace only

### Distribution

- MUST be one of: `kind`, `k3d`, `eks`
- Case-insensitive matching
- Invalid values result in exit code 3

### Source Directory

- MUST be existing directory or creatable path
- MUST have write permissions
- Relative paths resolved against current working directory

## Output Behavior

### Success Flow

1. Display spinner with "Initializing project..." message
2. Show file creation messages as files are generated:

   ```txt
   ✓ created 'ksail.yaml'
   ✓ created 'kind.yaml'
   ✓ created 'k8s/kustomization.yaml'
   ```

### Error Flow

- Clear error message explaining the problem
- Actionable suggestions for resolution
- Appropriate exit code
- No partial state left behind

### Progress Indication

- Spinner animation during operation
- Real-time file creation notifications
- Total operation time displayed on completion

## File Generation Contract

### Required Files

All projects MUST generate these files:

1. **ksail.yaml** - KSail configuration with specified distribution
2. **{distribution}.yaml** - Distribution-specific configuration (kind.yaml, k3d.yaml, etc.)
3. **k8s/kustomization.yaml** - Basic Kustomize structure

### Optional Files

Generated based on flags:

1. **Additional distribution configs** - Only when explicitly specified

### File Content Requirements

- All YAML files MUST be valid and parsable
- Configuration files MUST be compatible with target tools
- Generated content MUST pass validation by respective tools

## Error Scenarios

### Conflict Detection

When existing KSail files detected:

```bash
Error: KSail project already exists in this directory
Found: ksail.yaml, kind.yaml

Use --force to overwrite existing files or choose a different directory.
```

### Permission Errors

```bash
Error: Permission denied writing to directory '/path/to/dir'

Ensure you have write permissions or run with appropriate privileges.
```

### Validation Errors

```bash
Error: Invalid project name 'My-Project!'

Project name must be a valid DNS-1123 subdomain:
- Only lowercase letters, numbers, and hyphens
- Must start and end with alphanumeric character
- Maximum 63 characters
```

## Performance Requirements

- Total execution time: <5 seconds
- Memory usage: <50MB
- Responsive progress updates: <500ms intervals
- File I/O operations: <2 seconds total
