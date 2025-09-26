# Quickstart: KSail Init Command

This guide validates the complete user workflow for the KSail init command feature.

## Prerequisites

- KSail binary installed and in PATH
- Docker installed and running (for Kind clusters)
- Go 1.24+ development environment (for development testing)

## Basic Usage Test Scenario

### 1. Initialize a New Project

```bash
# Create and enter test directory
mkdir test-ksail-project
cd test-ksail-project

# Run init command with defaults
ksail init

# Expected output:
# ⠋ Initializing project...
# ✓ Created ksail.yaml
# ✓ Created kind.yaml
# ✓ Created k8s/kustomization.yaml
# ✓ Project initialized successfully!
#
# Next steps:
# 1. Run `ksail up` to create your cluster
# 2. Edit ksail.yaml to customize your configuration
# 3. Add your Kubernetes manifests to k8s/
```

**Validation**:
- Command completes in <5 seconds
- Three files created as shown
- Success message displays with next steps
- Exit code is 0

### 2. Verify Generated Files

```bash
# Check file structure
tree .
# Expected:
# .
# ├── ksail.yaml
# ├── kind.yaml
# └── k8s/
#     └── kustomization.yaml

# Validate ksail.yaml content
cat ksail.yaml
# Should contain:
# - kind distribution as default
# - valid YAML structure
# - project name matching directory

# Validate kind.yaml content
cat kind.yaml
# Should be valid Kind cluster configuration

# Validate kustomization.yaml
cat k8s/kustomization.yaml
# Should be valid Kustomize configuration
```

**Validation**:
- All files present and readable
- YAML syntax is valid
- Content matches expected templates

## Advanced Usage Test Scenario

### 3. Custom Configuration

```bash
# Clean previous test
cd .. && rm -rf test-ksail-project
mkdir custom-ksail-project
cd custom-ksail-project

# Run with custom options
ksail init --name my-cluster --distribution k3d --secret-manager

# Expected additional output:
# ✓ Created .sops.yaml
```

**Validation**:
- Custom name appears in ksail.yaml
- k3d.yaml created instead of kind.yaml
- .sops.yaml present when --secret-manager used
- Configuration reflects chosen options

### 4. Conflict Detection

```bash
# Try to init again without --force
ksail init

# Expected output:
# Error: KSail project already exists in this directory
# Found: ksail.yaml, k3d.yaml
#
# Use --force to overwrite existing files or choose a different directory.
```

**Validation**:
- Command fails with exit code 4
- Clear error message displayed
- No files modified
- Actionable suggestion provided

### 5. Force Overwrite

```bash
# Force overwrite with different options
ksail init --force --distribution kind

# Expected output shows overwriting
# ✓ Overwritten ksail.yaml
# ✓ Overwritten kind.yaml (replaced k3d.yaml)
```

**Validation**:
- Files successfully overwritten
- Configuration updated to new options
- Previous .sops.yaml removed if not specified

## Error Scenario Testing

### 6. Invalid Input Validation

```bash
# Test invalid project name
ksail init --name "Invalid Name!"

# Expected error:
# Error: Invalid project name 'Invalid Name!'
#
# Project name must be a valid DNS-1123 subdomain:
# - Only lowercase letters, numbers, and hyphens
# - Must start and end with alphanumeric character
# - Maximum 63 characters
```

### 7. Permission Testing

```bash
# Test read-only directory (Unix/Linux/macOS)
sudo chmod 444 .
ksail init --force

# Expected error:
# Error: Permission denied writing to directory '/path/to/dir'
#
# Ensure you have write permissions or run with appropriate privileges.

# Restore permissions
sudo chmod 755 .
```

## Integration Testing

### 8. Full Workflow Integration

```bash
# Initialize fresh project
cd .. && rm -rf custom-ksail-project
mkdir integration-test && cd integration-test

# Initialize with default settings
ksail init --name integration-test

# Verify files work with actual tools
kind create cluster --config kind.yaml --name integration-test

# Verify cluster creation
kubectl cluster-info --context kind-integration-test

# Cleanup
kind delete cluster --name integration-test
```

**Validation**:
- Generated kind.yaml works with kind CLI
- Cluster creates successfully
- KSail can connect to created cluster

## Performance Testing

### 9. Performance Validation

```bash
# Time the init operation
time ksail init --name perf-test

# Monitor memory usage during operation
# (Use appropriate monitoring tools for your platform)
```

**Validation**:
- Total time <5 seconds
- Memory usage <50MB during operation
- Responsive progress updates
- No memory leaks after completion

## Cleanup

```bash
# Remove test directories
cd ..
rm -rf test-ksail-project custom-ksail-project integration-test
```

## Success Criteria

This quickstart validates successful implementation when:

1. ✅ All basic usage scenarios complete successfully
2. ✅ Generated files are valid and functional
3. ✅ Error handling provides clear, actionable feedback
4. ✅ Performance meets specified requirements (<5s, <50MB)
5. ✅ Integration with existing KSail ecosystem works
6. ✅ Cross-platform compatibility verified
7. ✅ No regression in existing KSail functionality

## Troubleshooting

### Common Issues

**Command not found**: Ensure KSail binary is installed and in PATH
**Permission denied**: Check directory write permissions
**Invalid YAML**: Report as bug - generated content should always be valid
**Performance issues**: Check for I/O bottlenecks or memory constraints

### Debug Mode

```bash
# Enable verbose logging for debugging
KSAIL_LOG_LEVEL=debug ksail init --name debug-test
```

This quickstart serves as both user documentation and automated test validation for the KSail init command implementation.
