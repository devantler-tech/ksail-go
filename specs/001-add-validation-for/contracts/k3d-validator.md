# K3d Configuration Validator Contract

## Purpose
Validates k3d.yaml configuration files for K3d Kubernetes distribution compatibility and correctness.

## Validation Responsibilities

### K3d Schema Validation
- Validate k3d.yaml structure against K3d API v1alpha5.SimpleConfig schema
- Check required fields and proper configuration structure
- Validate server and agent node configurations
- Ensure registry and volume mapping correctness

### K3d-Specific Constraints
- Validate cluster name format and DNS compliance
- Check port mapping configurations and conflicts
- Validate registry mirror configurations
- Ensure volume mount specifications are accessible

### Server and Agent Configuration
- Validate server node specifications and resource limits
- Check agent node configurations and scaling settings
- Validate load balancer and ingress configurations
- Ensure network policy and security settings

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
