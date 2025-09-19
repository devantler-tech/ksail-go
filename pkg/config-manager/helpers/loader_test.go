package helpers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Name       string `yaml:"name"`
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// createDefaultConfig creates a default test configuration.
func createDefaultConfig() *testConfig {
	return &testConfig{Name: "default", APIVersion: "test/v1", Kind: "TestCluster"}
}

func TestLoadConfigFromFile(t *testing.T) {
	t.Parallel()

	t.Run("file exists", testLoadConfigFileExists)
	t.Run("file doesn't exist returns default", testLoadConfigFileNotExists)
	t.Run("invalid YAML returns error", testLoadConfigInvalidYAML)
}

func testLoadConfigFileExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	configContent := "name: test-cluster\napiVersion: test/v1\nkind: TestCluster"
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	config, err := helpers.LoadConfigFromFile(
		configPath,
		createDefaultConfig,
	)

	require.NoError(t, err)
	assert.Equal(t, "test-cluster", config.Name)
	assert.Equal(t, "test/v1", config.APIVersion)
	assert.Equal(t, "TestCluster", config.Kind)
}

func testLoadConfigFileNotExists(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non-existent.yaml")

	config, err := helpers.LoadConfigFromFile(
		configPath,
		createDefaultConfig,
	)

	require.NoError(t, err)
	assert.Equal(t, "default", config.Name)
	assert.Equal(t, "test/v1", config.APIVersion)
	assert.Equal(t, "TestCluster", config.Kind)
}

func testLoadConfigInvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0o600)
	require.NoError(t, err)

	_, err = helpers.LoadConfigFromFile(
		configPath,
		createDefaultConfig,
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}
