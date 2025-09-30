# pkg/provisioner/cluster/testutils

This package provides testing utilities for cluster provisioner testing.

## Purpose

Contains common test utilities, shared test cases, and helper functions for standardizing test patterns across cluster provisioner packages. This package provides testing infrastructure for Kind and K3d cluster provisioning functionality.

## Features

- **Common Test Cases**: Standardized test patterns for cluster provisioner testing
- **Shared Error Variables**: Common error types used across cluster provisioner tests
- **Helper Functions**: Utilities for setting up and tearing down test environments
- **Test Standardization**: Consistent testing patterns across different provisioner types
- **Error Scenario Testing**: Utilities for testing cluster creation and deletion failures

## Usage

```go
import clustertestutils "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/testutils"

func TestClusterProvisioner(t *testing.T) {
    // Use common error variables
    err := provisioner.CreateCluster()
    assert.Equal(t, clustertestutils.ErrCreateClusterFailed, err)

    // Use shared test patterns
    clustertestutils.RunCommonProvisionerTests(t, provisioner)
}
```

## Key Components

- **common.go**: Common test utilities and shared functionality
- **error variables**: Standardized error types for cluster operations
- **test patterns**: Reusable test cases for different provisioner types

## Integration

This package is used by cluster provisioner packages for testing:

- `pkg/provisioner/cluster/kind`: Kind cluster provisioning tests
- `pkg/provisioner/cluster/k3d`: K3d cluster provisioning tests

This ensures consistent testing patterns and reduces code duplication across all cluster provisioner functionality.

---

[⬅️ Go Back](../../../README.md)
