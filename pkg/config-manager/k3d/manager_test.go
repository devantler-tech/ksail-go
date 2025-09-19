package k3d_test

import (
	"testing"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers/testutils"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/k3d"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
)

// validateK3dDefaults validates K3d default configuration.
func validateK3dDefaults(t *testing.T, config *v1alpha5.SimpleConfig) {
	t.Helper()

	assert.Equal(t, "k3d.io/v1alpha5", config.APIVersion)
	assert.Equal(t, "Simple", config.Kind)
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
