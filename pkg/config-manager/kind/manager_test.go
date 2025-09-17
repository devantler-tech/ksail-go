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
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		content        string
		exists         bool
		expectError    bool
		expectedKind   string
		expectedAPIVer string
	}{
		{"non-existent file", "", false, false, "Cluster", "kind.x-k8s.io/v1alpha4"},
		{
			"valid config",
			"apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster",
			true,
			false,
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
		},
		{
			"missing TypeMeta",
			"nodes:\n- role: control-plane",
			true,
			false,
			"Cluster",
			"kind.x-k8s.io/v1alpha4",
		},
		{"invalid YAML", "invalid: [", true, true, "", ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")

			if testCase.exists {
				err := os.WriteFile(configPath, []byte(testCase.content), 0o600)
				require.NoError(t, err)
			}

			manager := kind.NewConfigManager(configPath)
			config, err := manager.LoadConfig()

			if testCase.expectError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)
			assert.Equal(t, testCase.expectedKind, config.Kind)
			assert.Equal(t, testCase.expectedAPIVer, config.APIVersion)
		})
	}
}

// TestLoadConfigCaching tests that LoadConfig properly caches results.
func TestLoadConfigCaching(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := "apiVersion: kind.x-k8s.io/v1alpha4\nkind: Cluster"
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)

	config1, err := manager.LoadConfig()
	require.NoError(t, err)

	config2, err := manager.LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, config1, config2)
}

// TestLoadConfigPathTraversal tests path traversal functionality.
func TestLoadConfigPathTraversal(t *testing.T) {
	t.Parallel()

	// Create a nested directory structure
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config.yaml")
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0o750)
	require.NoError(t, err)

	// Write config to parent directory
	configContent := `apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
name: traversal-test`
	err = os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	// Change to subdirectory for testing traversal
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		err := os.Chdir(oldDir)
		require.NoError(t, err)
	})
	err = os.Chdir(subDir)
	require.NoError(t, err)

	// Test with relative path - should find config in parent directory
	manager := kind.NewConfigManager("kind-config.yaml")
	config, err := manager.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
	assert.Equal(t, "traversal-test", config.Name)
}

// TestLoadConfigFileReadError tests error handling when file cannot be read.
func TestLoadConfigFileReadError(t *testing.T) {
	t.Parallel()

	// Create a directory instead of a file to trigger read error
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
