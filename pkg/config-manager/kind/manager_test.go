package kind_test

import (
	"testing"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers/testutils"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	"github.com/stretchr/testify/assert"
	v1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// validateKindDefaults validates Kind default configuration.
func validateKindDefaults(t *testing.T, config *v1alpha4.Cluster) {
	t.Helper()

	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
	assert.Equal(t, "Cluster", config.Kind)
}

// validateKindConfig validates Kind configuration with specific values.
func validateKindConfig(
	expectedName string,
	expectedNodeCount int,
) func(t *testing.T, config *v1alpha4.Cluster) {
	return func(t *testing.T, config *v1alpha4.Cluster) {
		t.Helper()

		validateKindDefaults(t, config)
		assert.Equal(t, expectedName, config.Name)
		assert.Len(t, config.Nodes, expectedNodeCount)
	}
}

func TestNewConfigManager(t *testing.T) {
	t.Parallel()

	configPath := "/path/to/config.yaml"
	manager := kind.NewConfigManager(configPath)

	assert.NotNil(t, manager)
}

// TestLoadConfig tests the LoadConfig method with different scenarios.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	scenarios := []testutils.TestScenario[v1alpha4.Cluster]{
		{
			Name:                "non-existent file",
			ConfigContent:       "",
			UseCustomConfigPath: false,
			ValidationFunc:      validateKindDefaults,
		},
		{
			Name: "valid config",
			ConfigContent: `apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
name: test-cluster
nodes:
- role: control-plane
- role: worker`,
			UseCustomConfigPath: true,
			ValidationFunc:      validateKindConfig("test-cluster", 2),
		},
		{
			Name: "missing TypeMeta",
			ConfigContent: `name: no-typemeta
nodes:
- role: control-plane`,
			UseCustomConfigPath: true,
			ValidationFunc:      validateKindConfig("no-typemeta", 1),
		},
		{
			Name:                "invalid YAML",
			ConfigContent:       "invalid: yaml: content: [",
			UseCustomConfigPath: true,
			ShouldError:         true,
		},
	}

	testutils.RunConfigManagerTests(
		t,
		func(configPath string) configmanager.ConfigManager[v1alpha4.Cluster] {
			return kind.NewConfigManager(configPath)
		},
		scenarios,
	)
}
