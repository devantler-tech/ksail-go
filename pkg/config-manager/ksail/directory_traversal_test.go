package ksail_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigManager_DirectoryTraversal tests the directory traversal functionality.
func TestConfigManager_DirectoryTraversal(t *testing.T) {
	// Create a temporary directory structure
	tempDir := t.TempDir()

	// Create parent directory with config file
	parentDir := filepath.Join(tempDir, "parent")
	err := os.MkdirAll(parentDir, 0o755)
	require.NoError(t, err)

	configContent := `
apiVersion: ksail.devantler.tech/v1alpha1
kind: Cluster
metadata:
  name: parent-config
spec:
  distribution: K3d
  sourceDirectory: parent-k8s
`

	configFile := filepath.Join(parentDir, "ksail.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err)

	// Create child directory
	childDir := filepath.Join(parentDir, "child")
	err = os.MkdirAll(childDir, 0o755)
	require.NoError(t, err)

	// Change to child directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(childDir)
	require.NoError(t, err)

	// Create config manager and load config
	manager := ksail.NewConfigManager()
	config, err := manager.LoadConfig()
	require.NoError(t, err)

	// Verify that the parent config was loaded
	assert.Equal(t, "parent-config", config.Metadata.Name)
	assert.Equal(t, "K3d", string(config.Spec.Distribution))
	assert.Equal(t, "parent-k8s", config.Spec.SourceDirectory)
}

// TestConfigManager_LoadConfigCaching tests that LoadConfig caches results.
func TestConfigManager_LoadConfigCaching(t *testing.T) {
	manager := ksail.NewConfigManager()

	// First call should load the config
	config1, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config1)

	// Second call should return the cached config (same pointer)
	config2, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config2)

	// Should be the same instance (cached)
	assert.Same(t, config1, config2)
}

// TestConfigManager_NoConfigFile tests behavior when no config file is found.
func TestConfigManager_NoConfigFile(t *testing.T) {
	// Create a temporary empty directory
	tempDir := t.TempDir()

	// Change to empty directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Create config manager and load config
	manager := ksail.NewConfigManager()
	config, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Should have default values
	assert.NotEmpty(t, config.TypeMeta.Kind)
	assert.NotEmpty(t, config.TypeMeta.APIVersion)
}
