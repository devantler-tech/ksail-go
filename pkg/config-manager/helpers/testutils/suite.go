// Package testutils provides common test utilities for config manager testing.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testFilePermissions = 0o600

// TestScenario represents a test scenario for config managers.
type TestScenario[T any] struct {
	Name                string
	ConfigContent       string
	APIVersion          string
	Kind                string
	ExpectedName        string
	ShouldError         bool
	ValidationFunc      func(t *testing.T, config *T)
	SetupFunc           func(t *testing.T) string // Returns config path
	UseCustomConfigPath bool
}

// RunConfigManagerTests runs a comprehensive test suite for config managers.
func RunConfigManagerTests[T any](
	t *testing.T,
	newManager func(configPath string) configmanager.ConfigManager[T],
	scenarios []TestScenario[T],
) {
	t.Helper()

	t.Run("basic scenarios", func(t *testing.T) {
		testLoadConfigBasicScenarios(t, newManager, scenarios)
	})
	t.Run("caching", func(t *testing.T) {
		var validScenario *TestScenario[T]

		for i := range scenarios {
			if !scenarios[i].ShouldError {
				validScenario = &scenarios[i]

				break
			}
		}

		require.NotNil(t, validScenario, "No valid scenario found for caching test")
		testLoadConfigCaching(t, newManager, *validScenario)
	})
}

// testLoadConfigBasicScenarios tests basic config loading scenarios.
func testLoadConfigBasicScenarios[T any](
	t *testing.T,
	newManager func(configPath string) configmanager.ConfigManager[T],
	scenarios []TestScenario[T],
) {
	t.Helper()
	t.Parallel()

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel()

			configPath := setupTestConfigPath(t, scenario)

			manager := newManager(configPath)
			config, err := manager.LoadConfig()

			if scenario.ShouldError {
				require.Error(t, err)
				assert.Nil(t, config)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, config)

			if scenario.ValidationFunc != nil {
				scenario.ValidationFunc(t, config)
			}
		})
	}
}

// setupTestConfigPath sets up the configuration path for a test scenario.
func setupTestConfigPath[T any](t *testing.T, scenario TestScenario[T]) string {
	t.Helper()

	switch {
	case scenario.SetupFunc != nil:
		return scenario.SetupFunc(t)
	case scenario.UseCustomConfigPath:
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.yaml")

		if scenario.ConfigContent != "" {
			err := os.WriteFile(configPath, []byte(scenario.ConfigContent), testFilePermissions)
			require.NoError(t, err)
		}

		return configPath
	default:
		return "non-existent-config.yaml"
	}
}

// testLoadConfigCaching tests that configuration caching works correctly.
func testLoadConfigCaching[T any](
	t *testing.T,
	newManager func(configPath string) configmanager.ConfigManager[T],
	scenario TestScenario[T],
) {
	t.Helper()
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "caching-config.yaml")

	if scenario.ConfigContent != "" {
		err := os.WriteFile(configPath, []byte(scenario.ConfigContent), testFilePermissions)
		require.NoError(t, err)
	}

	manager := newManager(configPath)

	// First call
	config1, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config1)

	// Second call should return the same instance (cached)
	config2, err := manager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config2)

	// Should be the same pointer (cached)
	assert.Same(t, config1, config2)
}
