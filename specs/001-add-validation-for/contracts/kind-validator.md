# Kind Configuration Validator Contract

## Purpose

Validates kind.yaml configuration files for Kind Kubernetes distribution compatibility and correctness **using official Kind upstream APIs**.

## ⚠️ CRITICAL IMPLEMENTATION REQUIREMENT

**MUST USE UPSTREAM KIND APIS**: This validator MUST leverage `sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster` for all validation logic to ensure complete compatibility with the Kind tool. Avoid custom validation that duplicates Kind's built-in validation functionality.

## Validation Responsibilities

### Kind Schema Validation (Using Upstream APIs)

- **UPSTREAM FIRST**: Use v1alpha4.Cluster struct unmarshaling for primary validation
- Validate kind.yaml structure against official Kind API v1alpha4.Cluster schema
- Leverage Kind's built-in validation methods where available
- Check required fields and proper nesting structure using Kind's validation
- Validate node configuration arrays and networking settings via Kind APIs
- Ensure image and version compatibility using Kind's validation logic

### Kind-Specific Constraints

- Validate cluster name format and constraints
- Check port mapping configurations for conflicts
- Validate volume mount paths and accessibility
- Ensure network configuration consistency

### Node Configuration Validation

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
