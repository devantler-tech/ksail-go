package kind_test

import (
	"os"
	"path/filepath"
	"testing"

	kind "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewConfigManager tests the NewConfigManager constructor.
func TestNewConfigManager(t *testing.T) {
	t.Parallel()

	configPath := "/tmp/test-kind-config.yaml"
	manager := kind.NewConfigManager(configPath)

	assert.NotNil(t, manager)
}

// TestLoadConfig tests the LoadConfig method with different scenarios.
//
// TestLoadConfig tests the LoadConfig method with different scenarios.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("basic scenarios", testLoadConfigBasicScenarios)
	t.Run("caching", testLoadConfigCaching)
	t.Run("path traversal", testLoadConfigPathTraversal)
	t.Run("file read error", testLoadConfigFileReadError)
}

// testLoadConfigBasicScenarios tests basic configuration loading scenarios.
func testLoadConfigBasicScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, content, expectedKind, expectedAPIVer string
		exists, expectError                         bool
	}{
		{"non-existent file", "", "Cluster", "kind.x-k8s.io/v1alpha4", false, false},
		{
			"valid config",
			"apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster",
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
			true,
			false,
		},
		{
			"missing TypeMeta",
			"nodes:\n- role: control-plane",
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
			true,
			false,
		},
		{"invalid YAML", "invalid: [", "", "", true, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")

			if tc.exists {
				err := os.WriteFile(configPath, []byte(tc.content), 0o600)
				require.NoError(t, err)
			}

			manager := kind.NewConfigManager(configPath)
			config, err := manager.LoadConfig()

			if tc.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Equal(t, tc.expectedKind, config.Kind)
			assert.Equal(t, tc.expectedAPIVer, config.APIVersion)
		})
	}
}

// testLoadConfigCaching tests configuration caching functionality.
func testLoadConfigCaching(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	content := "apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster"
	err := os.WriteFile(configPath, []byte(content), 0o600)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)
	config1, err := manager.LoadConfig()
	require.NoError(t, err)
	config2, err := manager.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, config1, config2)
}

// testLoadConfigPathTraversal tests path traversal functionality.
func testLoadConfigPathTraversal(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config.yaml")
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0o750)
	require.NoError(t, err)

	configContent := "apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster\nname: traversal-test"
	err = os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		err := os.Chdir(oldDir)
		require.NoError(t, err)
	})
	err = os.Chdir(subDir)
	require.NoError(t, err)

	manager := kind.NewConfigManager("kind-config.yaml")
	config, err := manager.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
	assert.Equal(t, "traversal-test", config.Name)
}

// testLoadConfigFileReadError tests file read error handling.
func testLoadConfigFileReadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config")
	err := os.Mkdir(configPath, 0o750)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)
	config, err := manager.LoadConfig()

	require.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "failed to read config file")
}
