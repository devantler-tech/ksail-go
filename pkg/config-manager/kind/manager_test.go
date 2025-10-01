package kind_test

import (
	"io"
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

// assertKindBasicConfig asserts basic configuration properties for Kind cluster.
func assertKindBasicConfig(t *testing.T, config *v1alpha4.Cluster, expectedName string) {
	t.Helper()

	assert.NotNil(t, config)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, expectedName, config.Name)
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

// TestNewKindCluster tests the NewKindCluster constructor.
func TestNewKindCluster(t *testing.T) {
	t.Parallel()

	t.Run("with_all_parameters", func(t *testing.T) {
		t.Parallel()

		config := kind.NewKindCluster(
			"test-cluster",
			"kind.x-k8s.io/v1alpha4",
			"Cluster",
		)

		assertKindBasicConfig(t, config, "test-cluster")
	})

	t.Run("with_empty_name", func(t *testing.T) {
		t.Parallel()

		config := kind.NewKindCluster(
			"",
			"kind.x-k8s.io/v1alpha4",
			"Cluster",
		)

		assert.NotNil(t, config)
		assert.Equal(t, "kind", config.Name)
	})

	t.Run("with_empty_apiVersion_and_kind", func(t *testing.T) {
		t.Parallel()

		config := kind.NewKindCluster(
			"test-cluster",
			"",
			"",
		)

		assertKindBasicConfig(t, config, "test-cluster")
	})

	t.Run("with_all_empty_values", func(t *testing.T) {
		t.Parallel()

		config := kind.NewKindCluster("", "", "")

		assertKindBasicConfig(t, config, "kind")
	})
}

func TestNewConfigManager(t *testing.T) {
	t.Parallel()

	configPath := "/path/to/config.yaml"
	manager := kind.NewConfigManager(configPath, io.Discard)

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
			return kind.NewConfigManager(configPath, io.Discard)
		},
		scenarios,
	)
}
