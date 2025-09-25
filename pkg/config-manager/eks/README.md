# pkg/config-manager/eks

This package provides configuration management for EKS cluster configurations.

## Purpose

Implements file-based configuration loading for EKS clusters using the `github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5.ClusterConfig` type. Provides simple YAML file loading with automatic eksctl defaults application and integrated validation.

## Features

- **File-based Loading**: Load EKS cluster configurations from YAML files
- **Default Configuration**: Returns sensible defaults when configuration file doesn't exist
- **Path Traversal**: For relative paths, searches up the directory tree to find configuration files
- **Caching**: Loads configuration once and caches for subsequent calls
- **EKS Defaults**: Automatically applies eksctl's built-in defaults via `eksctlapi.SetClusterConfigDefaults`
- **TypeMeta Completion**: Ensures proper APIVersion and Kind fields are set
- **Integrated Validation**: Validates configurations using the EKS validator after loading
- **Fail-fast Behavior**: Returns errors immediately if configuration validation fails

## Usage

```go
import (
    "github.com/devantler-tech/ksail-go/pkg/config-manager/eks"
)

// Create a config manager
manager := eks.NewConfigManager("eks.yaml")

// Load configuration (loads from file or returns default if file doesn't exist)
config, err := manager.LoadConfig()
if err != nil {
    // Handle error - this includes validation failures
    log.Fatal(err)
}

// Use the configuration
fmt.Printf("Cluster: %s in region %s\n", config.Metadata.Name, config.Metadata.Region)
```

## Default Configuration

When no configuration file exists, creates a default EKS cluster configuration with:

- **APIVersion**: `eksctl.io/v1alpha5`
- **Kind**: `ClusterConfig`
- **Name**: `default-cluster`
- **Region**: `us-west-2`
- **Standard Fields**: Initializes all required fields for a basic EKS cluster

## Configuration Validation

The config manager integrates with the EKS validator to ensure loaded configurations are valid:

- **Required Fields**: Validates cluster name and region are present and valid
- **Upstream Validation**: Uses eksctl's built-in validation for comprehensive checks
- **Error Messages**: Provides actionable error messages with fix suggestions
- **Fail-fast**: Returns errors immediately rather than allowing invalid configurations

## File Structure

- `manager.go`: Core configuration manager implementation
- `manager_test.go`: Comprehensive tests including validation scenarios
- `README.md`: This documentation file

## Dependencies

- `github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5`: EKS configuration types
- `github.com/devantler-tech/ksail-go/pkg/config-manager/helpers`: File loading utilities
- `github.com/devantler-tech/ksail-go/pkg/validator/eks`: EKS configuration validation

## Testing

Run tests with:

```bash
go test ./pkg/config-manager/eks/...
```

Tests cover:

- Configuration loading from files
- Default configuration generation
- Configuration validation integration
- Error handling for invalid configurations
- Caching behavior
