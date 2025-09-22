# EKS Configuration Validator Contract

## Implementation

### Dedicated EKS Validator for Loaded Structs

```go
type EKSValidator struct{}

// Validate performs validation on a loaded EKS cluster configuration
func (v *EKSValidator) Validate(config *EKSClusterConfig) *ValidationResult
```

## Purpose

Validates loaded EKS cluster configuration structs and ensures compatibility with AWS EKS requirements and KSail configuration settings. This validator operates on configuration structs that have already been loaded by the EKS config manager.

## 🔗 UPSTREAM API REQUIREMENT

**CRITICAL**: Use official EKS APIs from upstream Go packages wherever possible:

- `github.com/weaveworks/eksctl` - Official eksctl APIs for EKS configuration validation
- AWS SDK Go v2 packages for EKS service validation where needed
- **DO NOT** implement custom validation logic that duplicates functionality available in these official packages

The validation approach should:

1. Operate on loaded EKS configuration structs provided by the EKS config manager
2. Leverage eksctl's built-in validation methods where available on the loaded struct
3. Only add custom validation for KSail-specific cross-configuration requirements
4. Ensure compatibility with official AWS EKS tools and reduce maintenance burden

## Validation Responsibilities

### 🎯 **PRIMARY RESPONSIBILITY: EKS Configuration Semantic Validation**

**CRITICAL**: The EKS validator is **ONLY** responsible for validating individual EKS configuration semantics. Cross-configuration consistency is handled by the KSail validator.

- **UPSTREAM FIRST**: Validate loaded EKS configuration struct using eksctl's validation APIs
- Check semantic correctness of node groups, networking configurations, and IAM settings
- Leverage eksctl validation for AWS region names, availability zones, and instance types
- Validate Kubernetes version compatibility with EKS service using eksctl APIs

### ⚠️ **LIMITED SCOPE: No Cross-Configuration Validation**

This validator **DOES NOT** handle cross-configuration consistency. The KSail validator handles:

- Comparing EKS config with KSail config for naming consistency
- Ensuring EKS settings align with KSail distribution specifications
- Orchestrating validation across multiple configuration sources

### EKS-Specific Validation Only

- Use eksctl validation for AWS resource constraints and service limits
- Validate EKS-supported instance types and configurations using eksctl APIs
- Check IAM role and policy requirements using eksctl validation where available
- Ensure cluster endpoint access configurations are valid per AWS requirements

### Error Reporting and Remediation

- Provide specific error messages when loaded EKS config name mismatches KSail name
- Include field paths and suggested fixes for configuration inconsistencies in struct fields
- Reference AWS-specific requirements and constraints in error messages
- Format errors to clearly indicate whether issue is in EKS config semantics or cross-config consistency
- Provide guidance on AWS resource requirements and limitations for KSail compatibility

## Input/Output Contract

### Supported Configuration Types

- "eks" - EKS cluster configuration files (typically eksctl format)

### Validation Input

- Raw EKS configuration content as byte array
- OR parsed eksctl ClusterConfig struct

### Validation Output

- ValidationResult with overall status
- AWS-specific errors for resource constraints
- Actionable error messages for EKS configuration fixes
- Location information for error debugging

## Error Categories

### Schema Errors

- Missing required EKS fields
- Invalid AWS resource names
- Malformed cluster configuration
- Invalid Kubernetes version specifications

### AWS Resource Errors

- Invalid AWS regions or availability zones
- Unsupported EC2 instance types for EKS
- Invalid VPC or subnet configurations
- IAM permission or role issues
- Security group constraint violations

### EKS Service Errors

- Cluster name conflicts
- Service quota exceeded
- Incompatible add-on configurations
- Invalid endpoint access configurations

## Integration Points

### AWS SDK Integration

- Use AWS SDK for region and instance type validation
- Leverage eksctl APIs for configuration validation
- Validate against current AWS service quotas

### Config Manager Integration

- Integrate with existing pkg/config-manager loading logic
- Support both eksctl YAML and KSail EKS configurations
- Preserve existing EKS configuration loading patterns

### CLI Command Integration

- Validate EKS configurations during cluster operations
- Provide AWS-specific error reporting
- Support both human-readable and JSON output formats
