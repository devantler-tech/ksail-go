// Package testutils provides common test utilities for config manager testing.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
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
			err := manager.LoadConfig(nil)

			if scenario.ShouldError {
				require.Error(t, err)
				assert.Nil(t, manager.GetConfig())

				return
			}

			require.NoError(t, err)
			require.NotNil(t, manager.GetConfig())

			if scenario.ValidationFunc != nil {
				scenario.ValidationFunc(t, manager.GetConfig())
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

// AssertConfigManagerCaches verifies that a config manager reuses a previously loaded configuration
// when the underlying file becomes invalid after the initial load.
func AssertConfigManagerCaches[T any](
	t *testing.T,
	fileName string,
	configContent string,
	newManager func(configPath string) configmanager.ConfigManager[T],
) {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, fileName)

	err := os.WriteFile(configPath, []byte(configContent), testFilePermissions)
	require.NoError(t, err, "failed to write config")

	manager := newManager(configPath)

	err = manager.LoadConfig(nil)
	require.NoError(t, err, "initial LoadConfig failed")

	first := manager.GetConfig()
	require.NotNil(t, first, "expected config to be loaded")

	err = os.WriteFile(configPath, []byte("invalid: yaml: ["), testFilePermissions)
	require.NoError(t, err, "failed to overwrite config")

	err = manager.LoadConfig(nil)
	require.NoError(t, err, "expected cached load to succeed")

	require.Same(t, first, manager.GetConfig(), "expected cached configuration to be reused")
}
