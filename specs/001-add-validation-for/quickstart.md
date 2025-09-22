# Configuration File Validation - Quickstart Guide

## Overview

This guide demonstrates the configuration file validation feature that validates ksail.yaml, kind.yaml, k3d.yaml, and other configuration files automatically whenever they are loaded.

## Prerequisites

- KSail CLI installed and available in PATH
- Basic understanding of Kubernetes configuration
- Sample configuration files for testing

## Quick Validation Test

### Test 1: Valid Configuration Validation

Create a valid ksail.yaml file and verify validation passes:

```bash
# Create a valid ksail.yaml configuration
cat > ksail.yaml << 'EOF'
apiVersion: cluster.ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  connection:
    context: kind-test-cluster
EOF

# Create corresponding valid kind.yaml
cat > kind.yaml << 'EOF'
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
name: test-cluster
nodes:
- role: control-plane
- role: worker
EOF

# Run any ksail command to trigger validation
ksail status
```

**Expected Result**: Command executes successfully without validation errors.

### Test 2: Invalid YAML Syntax Validation

Test that malformed YAML is caught with actionable error messages:

```bash
# Create invalid YAML with syntax error
cat > ksail.yaml << 'EOF'
apiVersion: cluster.ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
  # Missing closing quote
  description: "Invalid YAML
spec:
  distribution: Kind
EOF

# Attempt to run ksail command
ksail status
```

**Expected Result**: Clear error message indicating YAML syntax error with line number and fix suggestion.

### Test 3: Invalid Field Values Validation

Test semantic validation with invalid field values:

```bash
# Create YAML with invalid distribution value
cat > ksail.yaml << 'EOF'
apiVersion: cluster.ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
spec:
  distribution: InvalidDistribution
  distributionConfig: kind.yaml
EOF

# Attempt to run ksail command
ksail status
```

**Expected Result**: Validation error listing allowed distribution values (Kind, K3d, EKS) with example.

### Test 4: Missing Required Fields Validation

Test that missing required fields are detected:

```bash
# Create YAML missing required fields
cat > ksail.yaml << 'EOF'
apiVersion: cluster.ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
# Missing spec section entirely
EOF

# Attempt to run ksail command
ksail status
```

**Expected Result**: Error listing all missing required fields with examples.

### Test 5: Cross-Configuration Validation

Test validation of relationships between configuration files:

```bash
# Create ksail.yaml with kind distribution
cat > ksail.yaml << 'EOF'
apiVersion: cluster.ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
spec:
  distribution: Kind
  distributionConfig: kind.yaml
  connection:
    context: kind-test-cluster
EOF

# Create kind.yaml with mismatched cluster name
cat > kind.yaml << 'EOF'
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
name: different-cluster-name
nodes:
- role: control-plane
EOF

# Attempt to run ksail command
ksail status
```

**Expected Result**: Error indicating cluster name mismatch between ksail.yaml and kind.yaml with fix suggestion.

### Test 6: Performance Validation

Verify that validation completes quickly for typical configurations:

```bash
# Create valid configuration
cat > ksail.yaml << 'EOF'
apiVersion: cluster.ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: performance-test
spec:
  distribution: Kind
  distributionConfig: kind.yaml
EOF

# Time the validation (embedded in status command)
time ksail status
```

**Expected Result**: Command completes in well under 1 second, with validation taking minimal time.

## Error Message Examples

### YAML Syntax Error

```
Error: Configuration validation failed
File: /path/to/ksail.yaml:6:25
Field: metadata.description
Error: Invalid YAML syntax - unterminated quoted string
Current: "Invalid YAML
Expected: "Invalid YAML"
Fix: Add closing quote to complete the string value
```

### Invalid Field Value

```
Error: Configuration validation failed
File: /path/to/ksail.yaml:7:17
Field: spec.distribution
Error: Invalid distribution type
Current: "InvalidDistribution"
Expected: One of [Kind, K3d, EKS]
Fix: Change to a supported distribution, e.g., "Kind"
```

### Missing Required Field

```
Error: Configuration validation failed
File: /path/to/ksail.yaml
Missing required fields:
- spec.distribution (required): Kubernetes distribution type
- spec.distributionConfig (required): Path to distribution configuration file
Fix: Add missing fields to your ksail.yaml configuration
```

### Cross-Configuration Error

```
Error: Configuration validation failed
Field: kind.yaml cluster name
Error: Cluster name mismatch between configurations
Current: kind.yaml name "different-cluster-name"
Expected: Should match ksail.yaml metadata.name "test-cluster"
Fix: Change kind.yaml name to "test-cluster" or update ksail.yaml metadata.name
```

## Validation Integration Points

### Command Integration

Validation runs automatically during:

- `ksail init` - Validates generated configurations
- `ksail up` - Validates before cluster creation
- `ksail status` - Validates before status check
- `ksail down` - Validates before cluster deletion
- All other commands that load configuration

### Performance Characteristics

- Validation completes within 100ms for typical files
- Memory usage stays under 10MB during validation
- No temporary files or I/O operations during validation
- Suitable for frequent execution without performance impact

## Troubleshooting Common Issues

### "Configuration file not found"

**Solution**: Ensure ksail.yaml exists in current directory or run from project root.

### "Permission denied reading configuration"

**Solution**: Check file permissions and ensure files are readable.

### "Invalid YAML syntax"

**Solution**: Use YAML validator or IDE with YAML support to identify syntax issues.

### "Distribution config file not found"

**Solution**: Ensure the distributionConfig file path in ksail.yaml points to existing file.

## Success Criteria Verification

✅ **Immediate Feedback**: Validation errors appear immediately when commands are run
✅ **Actionable Messages**: Error messages include specific field paths and fix suggestions  
✅ **Fast Performance**: Validation completes well under 100ms for typical configurations
✅ **Comprehensive Coverage**: All configuration types (ksail, kind, k3d) are validated
✅ **Fail-Fast Behavior**: Invalid configurations prevent command execution
✅ **Cross-Configuration Validation**: Related configurations are checked for consistency
