# EKS Configuration Validator Contract

## Purpose
Validates EKS cluster configuration files and ensures compatibility with AWS EKS requirements.

## 🔗 UPSTREAM API REQUIREMENT
**CRITICAL**: Use official EKS APIs from upstream Go packages wherever possible:
- `github.com/weaveworks/eksctl` - Official eksctl APIs for EKS configuration validation
- AWS SDK Go v2 packages for EKS service validation
- **DO NOT** implement custom validation logic that duplicates functionality available in these official packages

This ensures compatibility with official AWS EKS tools and reduces maintenance burden.

## Validation Responsibilities

### EKS Configuration Schema Validation
- Validate EKS cluster configuration file structure
- Check required fields: metadata.name, metadata.region, nodeGroups
- Validate AWS region names against official AWS region list
- Validate instance types against EKS-supported instance types
- Validate Kubernetes version compatibility with EKS

### AWS Resource Validation
- Validate VPC and subnet configurations
- Check IAM role and policy requirements
- Validate security group configurations
- Ensure EKS service limits are not exceeded
- Validate node group scaling configurations

### EKS-Specific Constraints
- Validate managed node group configurations
- Check Fargate profile configurations (if applicable)
- Validate EKS add-on compatibility
- Ensure cluster endpoint access configurations are valid
- Validate logging configurations

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
