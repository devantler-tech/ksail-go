# Kind Configuration Validator Contract

## Implementation

### Dedicated Kind Validator for Loaded Structs

```go
type KindValidator struct{}

// Validate performs validation on a loaded Kind cluster configuration
func (v *KindValidator) Validate(config *v1alpha4.Cluster) *ValidationResult
```

## Purpose

Validates loaded v1alpha4.Cluster structs for Kind Kubernetes distribution compatibility and correctness **using official Kind upstream APIs**. This validator focuses on Kind configuration semantic validation and ensuring compatibility with KSail configuration settings, operating on structs that have already been loaded by the Kind config manager.

## ⚠️ CRITICAL IMPLEMENTATION REQUIREMENT

**MUST USE UPSTREAM KIND APIS**: This validator MUST leverage `sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster` for all validation logic to ensure complete compatibility with the Kind tool. The validation approach should:

1. Operate on loaded `v1alpha4.Cluster` structs provided by the Kind config manager
2. Leverage Kind's built-in validation methods where available on the loaded struct
3. Only add custom validation for KSail-specific cross-configuration requirements
4. Avoid duplicating validation that Kind config manager and APIs already provide

## Validation Responsibilities

### 🎯 **PRIMARY RESPONSIBILITY: Kind Configuration Semantic Validation**

**CRITICAL**: The Kind validator is **ONLY** responsible for validating individual Kind configuration semantics. Cross-configuration consistency is handled by the KSail validator.

- **UPSTREAM FIRST**: Validate loaded `v1alpha4.Cluster` struct using Kind's validation APIs
- Check semantic correctness of node configurations, networking settings, and volume mounts
- Leverage Kind's built-in validation methods for field constraints and dependencies
- Validate Kind-specific features like extra port mappings and container image settings

### ⚠️ **LIMITED SCOPE: No Cross-Configuration Validation**

This validator **DOES NOT** handle cross-configuration consistency. The KSail validator handles:

- Comparing Kind config with KSail config for naming consistency
- Ensuring Kind settings align with KSail distribution specifications
- Orchestrating validation across multiple configuration sources

### Kind-Specific Validation Only

- Validate control-plane and worker node configurations
- Check resource allocations and limits
- Validate extra mounts and port mappings
- Ensure container image specifications are valid

## Input/Output Contract

### Supported Configuration Types

- "kind" - kind.yaml configuration files

### Validation Input

- Raw kind.yaml content as byte array
- OR parsed v1alpha4.Cluster struct

### Validation Output

- ValidationResult with Kind-specific validation status
- Detailed errors for Kind configuration issues
- Actionable remediation suggestions
- Location information for error sources

## Error Categories

### Schema Errors

- Invalid Kind API version or kind fields
- Missing required configuration sections
- Malformed node configuration arrays
- Invalid field types or structures

### Configuration Errors

- Invalid cluster naming (non-DNS compliant)
- Port mapping conflicts or invalid ranges
- Invalid image references or tags
- Resource constraint violations

### Networking Errors

- Invalid subnet or IP range specifications
- Port conflicts between services
- Invalid ingress controller configurations
- DNS configuration issues

### Volume and Mount Errors

- Invalid host path specifications
- Inaccessible mount points
- Permission issues with volume mounts
- Invalid container path specifications

## Independence Requirements

### No External Dependencies

- Must NOT depend on ksail validator or other validators
- Must NOT require external configuration files
- Must operate solely on provided Kind configuration data
- Must NOT perform file system operations for validation

### Focused Validation Scope

- Only validates Kind-specific configuration elements
- Does NOT validate cross-configuration consistency
- Does NOT check ksail.yaml compatibility
- Focuses solely on Kind API schema compliance

## Integration Contract

### Standalone Operation

- Can be used independently of other validators
- Provides complete validation for Kind configurations
- Returns comprehensive validation results
- Supports direct testing and usage

### Called by KSail Validator

- Invoked by ksail validator for Kind distribution configs
- Receives parsed Kind configuration data
- Returns validation results for aggregation
- Maintains independence while supporting coordination
