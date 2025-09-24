package eks_test

import (
	"testing"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/eks"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers/testutils"
	"github.com/stretchr/testify/assert"
	eksctlapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// validateEKSConfigStructure validates the basic EKS configuration structure.
func validateEKSConfigStructure(
	t *testing.T,
	config *eksctlapi.ClusterConfig,
	expectedName, expectedRegion string,
) {
	t.Helper()

	assert.Equal(t, "eksctl.io/v1alpha5", config.APIVersion)
	assert.Equal(t, "ClusterConfig", config.Kind)
	assert.NotNil(t, config.Metadata)
	assert.Equal(t, expectedName, config.Metadata.Name)
	assert.Equal(t, expectedRegion, config.Metadata.Region)
}

// validateEKSDefaults validates EKS default configuration.
func validateEKSDefaults(t *testing.T, config *eksctlapi.ClusterConfig) {
	t.Helper()
	validateEKSConfigStructure(t, config, "eks-default", "eu-north-1")
}

// validateEKSConfig validates EKS configuration with specific values.
func validateEKSConfig(
	expectedName string,
	expectedRegion string,
) func(t *testing.T, config *eksctlapi.ClusterConfig) {
	return func(t *testing.T, config *eksctlapi.ClusterConfig) {
		t.Helper()
		validateEKSConfigStructure(t, config, expectedName, expectedRegion)
	}
}

func TestNewConfigManager(t *testing.T) {
	t.Parallel()

	configPath := "/path/to/config.yaml"
	manager := eks.NewConfigManager(configPath)

	assert.NotNil(t, manager)
}

// TestLoadConfig tests the LoadConfig method with different scenarios.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	scenarios := []testutils.TestScenario[eksctlapi.ClusterConfig]{
		{
			Name:                "non-existent file",
			ConfigContent:       "",
			UseCustomConfigPath: false,
			ValidationFunc:      validateEKSDefaults,
		},
		{
			Name: "valid EKS config",
			ConfigContent: `apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: test-cluster
  region: us-east-1
`,
			UseCustomConfigPath: true,
			ValidationFunc:      validateEKSConfig("test-cluster", "us-east-1"),
		},
		{
			Name: "invalid EKS config - empty name",
			ConfigContent: `apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: ""
  region: us-east-1
`,
			UseCustomConfigPath: true,
			ShouldError:         true,
		},
		{
			Name: "invalid EKS config - empty region",
			ConfigContent: `apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: test-cluster
  region: ""
`,
			UseCustomConfigPath: true,
			ShouldError:         true,
		},
	}

	testutils.RunConfigManagerTests(
		t,
		func(configPath string) configmanager.ConfigManager[eksctlapi.ClusterConfig] {
			return eks.NewConfigManager(configPath)
		},
		scenarios,
	)
}

// TestNewEKSClusterConfig tests the NewEKSClusterConfig constructor.
func TestNewEKSClusterConfig(t *testing.T) {
	t.Parallel()

	config := eks.NewEKSClusterConfig(
		"test-cluster",
		"us-west-2",
		"eksctl.io/v1alpha5",
		"ClusterConfig",
	)

	assert.NotNil(t, config)
	assert.Equal(t, "eksctl.io/v1alpha5", config.APIVersion)
	assert.Equal(t, "ClusterConfig", config.Kind)
	assert.Equal(t, "test-cluster", config.Metadata.Name)
	assert.Equal(t, "us-west-2", config.Metadata.Region)
}
