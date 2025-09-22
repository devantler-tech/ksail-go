# K3d Configuration Validator Contract

## Implementation

### Dedicated K3d Validator for Loaded Structs

```go
type K3dValidator struct{}

// Validate performs validation on a loaded K3d cluster configuration
func (v *K3dValidator) Validate(config *v1alpha5.SimpleConfig) *ValidationResult
```

## Purpose

Validates loaded v1alpha5.SimpleConfig structs for K3d Kubernetes distribution compatibility and correctness **using official K3d upstream APIs**. This validator focuses on K3d configuration semantic validation and ensuring compatibility with KSail configuration settings, operating on structs that have already been loaded by the K3d config manager.

## ⚠️ CRITICAL IMPLEMENTATION REQUIREMENT

**MUST USE UPSTREAM K3D APIS**: This validator MUST leverage `github.com/k3d-io/k3d/v5/pkg/config/v1alpha5.SimpleConfig` for all validation logic to ensure complete compatibility with the K3d tool. The validation approach should:

1. Operate on loaded `v1alpha5.SimpleConfig` structs provided by the K3d config manager
2. Leverage K3d's built-in validation methods where available on the loaded struct
3. Only add custom validation for KSail-specific cross-configuration requirements
4. Avoid duplicating validation that K3d config manager and APIs already provide

## Validation Responsibilities

### 🎯 **PRIMARY RESPONSIBILITY: K3d Configuration Semantic Validation**

**CRITICAL**: The K3d validator is **ONLY** responsible for validating individual K3d configuration semantics. Cross-configuration consistency is handled by the KSail validator.

- **UPSTREAM FIRST**: Validate loaded `v1alpha5.SimpleConfig` struct using K3d's validation APIs
- Check semantic correctness of server/agent configurations, networking options, and volume mappings
- Leverage K3d's built-in validation methods for field constraints and dependencies
- Validate K3d-specific features like registry mirrors and K3s extra arguments

### ⚠️ **LIMITED SCOPE: No Cross-Configuration Validation**

This validator **DOES NOT** handle cross-configuration consistency. The KSail validator handles:

- Comparing K3d config with KSail config for naming consistency
- Ensuring K3d settings align with KSail distribution specifications
- Orchestrating validation across multiple configuration sources

### K3d-Specific Validation Only

- Format errors to clearly indicate whether issue is in K3d config semantics or cross-config consistency
- Provide concrete examples of required K3d configuration changes for KSail compatibility

## Input/Output Contract

### Supported Configuration Types

- "k3d" - k3d.yaml configuration files

### Validation Input

- Raw k3d.yaml content as byte array
- OR parsed v1alpha5.SimpleConfig struct

### Validation Output

- ValidationResult with K3d-specific validation status
- Detailed errors for K3d configuration issues
- Actionable remediation suggestions
- Location information for error sources

## Error Categories

### Schema Errors

- Invalid K3d API version or configuration format
- Missing required configuration sections
- Malformed server or agent configurations
- Invalid field types or nested structures

### Configuration Errors

- Invalid cluster naming conventions
- Port mapping conflicts or invalid specifications
- Invalid image registry configurations
- Resource allocation issues

### Networking Errors

- Invalid network configurations
- Port conflicts between services and load balancers
- Invalid ingress controller settings
- DNS and service discovery issues

### Registry and Storage Errors

- Invalid registry mirror configurations
- Inaccessible volume mount paths
- Permission issues with storage mounts
- Invalid registry authentication settings

## Independence Requirements

### No External Dependencies

- Must NOT depend on ksail validator or other validators
- Must NOT require external configuration files
- Must operate solely on provided K3d configuration data
- Must NOT perform file system operations for validation

### Focused Validation Scope

- Only validates K3d-specific configuration elements
- Does NOT validate cross-configuration consistency
- Does NOT check ksail.yaml compatibility
- Focuses solely on K3d API schema compliance

## Integration Contract

### Standalone Operation

- Can be used independently of other validators
- Provides complete validation for K3d configurations
- Returns comprehensive validation results
- Supports direct testing and usage

### Called by KSail Validator

- Invoked by ksail validator for K3d distribution configs
- Receives parsed K3d configuration data
- Returns validation results for aggregation
- Maintains independence while supporting coordination
