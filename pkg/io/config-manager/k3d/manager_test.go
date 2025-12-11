package k3d_test

import (
	"testing"

	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"

	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	"github.com/devantler-tech/ksail-go/pkg/io/config-manager/testutils"
)

// validateK3dDefaults validates K3d default configuration.
func validateK3dDefaults(t *testing.T, config *v1alpha5.SimpleConfig) {
	t.Helper()

	assert.Equal(t, "k3d.io/v1alpha5", config.APIVersion)
	assert.Equal(t, "Simple", config.Kind)
}

// assertK3dBasicConfig asserts basic configuration properties for K3d cluster.
func assertK3dBasicConfig(t *testing.T, config *v1alpha5.SimpleConfig, expectedName string) {
	t.Helper()

	assert.NotNil(t, config)
	assert.Equal(t, "k3d.io/v1alpha5", config.APIVersion)
	assert.Equal(t, "Simple", config.Kind)
	assert.Equal(t, expectedName, config.Name)
}

// validateK3dConfig validates K3d configuration with specific values.
func validateK3dConfig(
	expectedName string,
	expectedServers, expectedAgents int,
) func(t *testing.T, config *v1alpha5.SimpleConfig) {
	return func(t *testing.T, config *v1alpha5.SimpleConfig) {
		t.Helper()

		validateK3dDefaults(t, config)
		assert.Equal(t, expectedName, config.Name)
		assert.Equal(t, expectedServers, config.Servers)
		assert.Equal(t, expectedAgents, config.Agents)
	}
}

// TestNewK3dSimpleConfig tests the NewK3dSimpleConfig constructor.
func TestNewK3dSimpleConfig(t *testing.T) {
	t.Parallel()

	t.Run("with_all_parameters", func(t *testing.T) {
		t.Parallel()

		config := k3d.NewK3dSimpleConfig(
			"test-cluster",
			"k3d.io/v1alpha5",
			"Simple",
		)

		assertK3dBasicConfig(t, config, "test-cluster")
	})

	t.Run("with_empty_name", func(t *testing.T) {
		t.Parallel()

		config := k3d.NewK3dSimpleConfig(
			"",
			"k3d.io/v1alpha5",
			"Simple",
		)

		assert.NotNil(t, config)
		assert.Equal(t, "k3d-default", config.Name)
	})

	t.Run("with_empty_apiVersion_and_kind", func(t *testing.T) {
		t.Parallel()

		config := k3d.NewK3dSimpleConfig(
			"test-cluster",
			"",
			"",
		)

		assertK3dBasicConfig(t, config, "test-cluster")
	})

	t.Run("with_all_empty_values", func(t *testing.T) {
		t.Parallel()

		config := k3d.NewK3dSimpleConfig("", "", "")

		assert.NotNil(t, config)
		assert.Equal(t, "k3d.io/v1alpha5", config.APIVersion)
		assert.Equal(t, "Simple", config.Kind)
		assert.Equal(t, "k3d-default", config.Name)
	})
}

func TestNewConfigManager(t *testing.T) {
	t.Parallel()

	configPath := "/path/to/config.yaml"
	manager := k3d.NewConfigManager(configPath)

	assert.NotNil(t, manager)
}

// TestLoadConfig tests the LoadConfig method with different scenarios.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	scenarios := []testutils.TestScenario[v1alpha5.SimpleConfig]{
		{
			Name:                "non-existent file",
			ConfigContent:       "",
			UseCustomConfigPath: false,
			ValidationFunc:      validateK3dDefaults,
		},
		{
			Name: "valid config",
			ConfigContent: `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: test-cluster
servers: 1
agents: 2`,
			UseCustomConfigPath: true,
			ValidationFunc:      validateK3dConfig("test-cluster", 1, 2),
		},
		{
			Name: "missing TypeMeta",
			ConfigContent: `metadata:
  name: no-typemeta
servers: 3`,
			UseCustomConfigPath: true,
			ValidationFunc:      validateK3dConfig("no-typemeta", 3, 0),
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
		func(configPath string) configmanager.ConfigManager[v1alpha5.SimpleConfig] {
			return k3d.NewConfigManager(configPath)
		},
		scenarios,
	)
}

func TestK3dConfigManagerLoadConfig_ReusesExistingConfig(t *testing.T) {
	t.Parallel()

	testutils.AssertConfigManagerCaches[v1alpha5.SimpleConfig](
		t,
		"k3d.yaml",
		`apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: cached
`,
		func(configPath string) configmanager.ConfigManager[v1alpha5.SimpleConfig] {
			return k3d.NewConfigManager(configPath)
		},
	)
}
