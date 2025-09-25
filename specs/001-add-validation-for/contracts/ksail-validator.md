# KSail Configuration Validator Contract

## Implementation

### KSail Validator for Loaded Structs

```go
type KSailValidator struct{}

// Validate performs validation on a loaded KSail cluster configuration
func (v *KSailValidator) Validate(config *v1alpha1.Cluster) *ValidationResult
```

## Purpose

The KSailValidator validates loaded v1alpha1.Cluster structs for semantic correctness and cross-configuration consistency. It operates on configuration structs that have already been loaded by the KSail config manager, focusing purely on validation logic without file handling concerns.

## ⚠️ CRITICAL CONSTRAINT

**DO NOT ALTER THE KSAIL CONFIG**: This validator must preserve the existing v1alpha1.Cluster API structure and all current configuration fields. The ksail configuration format is considered stable and must not be modified as part of this validation implementation.

## Validation Responsibilities

### KSail Configuration Schema Validation

- Validate loaded v1alpha1.Cluster struct fields against schema constraints
- Check required fields: metadata.name, spec.distribution, spec.distributionConfig
- Validate enum values for spec.distribution (Kind, K3d, EKS)
- Ensure field values are within valid ranges and follow expected patterns

### 🎯 **PRIMARY RESPONSIBILITY: Cross-Configuration Coordination**

**CRITICAL**: The KSail validator is the **ONLY** validator responsible for cross-configuration consistency. Distribution validators (Kind, K3d, EKS) handle only their own configurations.

- Coordinate with other config managers to load distribution-specific configurations
- Compare loaded KSail configuration with loaded distribution configurations
- Ensure naming consistency between loaded ksail and distribution config structs
- Validate that loaded configurations are mutually compatible
- **Orchestrate** validation by calling distribution validators for their specific configs
- **Aggregate** validation results from all configuration sources

### Context Name Validation

- Validate spec.connection.context field matches expected pattern for distribution:
  - Kind: context must be "kind-{metadata.name}"
  - K3d: context must be "k3d-{metadata.name}"
  - EKS: context must match AWS EKS cluster ARN or cluster name pattern
- Check context name patterns against loaded metadata.name field

### Configuration Consistency Validation

- Ensure loaded KSail settings are compatible with loaded distribution configurations
- Validate CNI settings alignment between ksail and distribution config structs
- Check CSI and ingress controller settings for consistency across loaded configs
- Verify that loaded configuration combinations are practically deployable## Input/Output Contract

### Supported Configuration Types

- "ksail" - ksail.yaml configuration files

### Validation Input

- Raw ksail.yaml content as byte array
- OR parsed v1alpha1.Cluster struct

### Validation Output

- ValidationResult with overall status
- Specific errors for schema violations
- Actionable error messages for field corrections
- Location information for error debugging

## Error Categories

### Schema Errors

- Missing required fields
- Invalid field types
- Enum constraint violations
- Malformed YAML syntax

### Semantic Errors

- Invalid distribution configuration references
- Context naming inconsistencies
- Unsupported distribution types
- Configuration compatibility issues

### Cross-Configuration Errors

- Distribution config file not found
- Naming mismatches between configurations
- Distribution-specific validation failures

## Integration Points

### Distribution Validator Coordination

- Call kind.Validator for Kind distribution configs
- Call k3d.Validator for K3d distribution configs
- Call eks.Validator for EKS distribution configs

### Config Manager Integration

- Integrate with existing pkg/config-manager loading logic
- Fail-fast validation during configuration loading
- Preserve existing configuration loading patterns

### CLI Command Integration

- Validate configurations during all command executions
- Provide consistent error reporting across commands
- Support both human-readable and JSON output formats
