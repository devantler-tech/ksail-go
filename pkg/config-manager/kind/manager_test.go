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

// TestLoadConfigWithNonExistentFile tests LoadConfig with missing file returns defaults.
func TestLoadConfigWithNonExistentFile(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non-existent.yaml")

	manager := kind.NewConfigManager(configPath)
	config, err := manager.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)

	// Test that subsequent calls return the same config (caching)
	config2, err2 := manager.LoadConfig()
	require.NoError(t, err2)
	assert.Equal(t, config, config2)
}

// TestLoadConfigWithValidFile tests LoadConfig with valid configuration file.
func TestLoadConfigWithValidFile(t *testing.T) {
	t.Parallel()

	configContent := `apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
- role: control-plane
- role: worker`

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config.yaml")

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)
	config, err := manager.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
}

// TestLoadConfigWithMinimalFile tests LoadConfig with minimal configuration.
func TestLoadConfigWithMinimalFile(t *testing.T) {
	t.Parallel()

	configContent := `apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster`

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config.yaml")

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)
	config, err := manager.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
}

// TestLoadConfigWithMissingTypeMeta tests LoadConfig fills in missing TypeMeta.
func TestLoadConfigWithMissingTypeMeta(t *testing.T) {
	t.Parallel()

	configContent := `nodes:
- role: control-plane`

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config.yaml")

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)
	config, err := manager.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
}

// TestLoadConfigWithInvalidYAML tests LoadConfig with invalid YAML returns error.
func TestLoadConfigWithInvalidYAML(t *testing.T) {
	t.Parallel()

	configContent := `apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
  - role: control-plane
	  invalid indentation`

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config.yaml")

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)
	config, err := manager.LoadConfig()

	require.Error(t, err)
	assert.Nil(t, config)
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

// TestLoadConfigDefaults tests that Kind defaults are properly applied.
func TestLoadConfigDefaults(t *testing.T) {
	t.Parallel()

	// Test with non-existent file (should get defaults)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non-existent.yaml")

	manager := kind.NewConfigManager(configPath)
	config, err := manager.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify Kind defaults are applied
	assert.Equal(t, "Cluster", config.Kind)
	assert.Equal(t, "kind.x-k8s.io/v1alpha4", config.APIVersion)
	// Kind defaults should include at least one control-plane node
	assert.GreaterOrEqual(t, len(config.Nodes), 1)
}

// TestLoadConfigCaching tests that LoadConfig properly caches results.
func TestLoadConfigCaching(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "kind-config.yaml")

	configContent := `apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
- role: control-plane`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	manager := kind.NewConfigManager(configPath)

	// First call should load from file
	config1, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config1)

	// Modify the file
	modifiedContent := `apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
nodes:
- role: control-plane
- role: worker`

	err = os.WriteFile(configPath, []byte(modifiedContent), 0o600)
	require.NoError(t, err)

	// Second call should return cached result (not re-read file)
	config2, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config2)

	// Should be same object (cached)
	assert.Equal(t, config1, config2)
	// Should not have picked up the file changes (proving caching works)
	assert.Len(t, config2.Nodes, 1) // Still has original single node
}
