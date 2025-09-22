# KSail Configuration Validator Contract

## Purpose

Validates ksail.yaml configuration files and coordinates validation of related distribution-specific configurations.

## ⚠️ CRITICAL CONSTRAINT

**DO NOT ALTER THE KSAIL CONFIG**: This validator must preserve the existing v1alpha1.Cluster API structure and all current configuration fields. The ksail configuration format is considered stable and must not be modified as part of this validation implementation.

## Validation Responsibilities

### KSail Configuration Schema Validation

- Validate ksail.yaml structure against cluster API v1alpha1 schema
- Check required fields: metadata.name, spec.distribution, spec.distributionConfig
- Validate enum values for spec.distribution (Kind, K3d, EKS)
- Validate spec.connection.context format consistency

### Cross-Configuration Coordination

- Load and validate distribution-specific configurations when needed
- Ensure naming consistency between ksail.yaml and distribution configs
- Validate that distribution config files exist and are accessible
- Coordinate with appropriate distribution validators

### Context Validation

- Validate spec.connection.context matches expected pattern for distribution
- For Kind: context must be "kind-{metadata.name}"
- For K3d: context must be "k3d-{metadata.name}"
- For EKS: context must match AWS EKS cluster ARN or cluster name pattern

## Input/Output Contract

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
