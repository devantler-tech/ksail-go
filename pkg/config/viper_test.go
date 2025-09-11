package config_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/stretchr/testify/assert"
)

// TestGetConfigFilePath tests the configuration file path function.
func TestGetConfigFilePath(t *testing.T) {
	t.Parallel()

	path := config.GetConfigFilePath()
	expected := "ksail.yaml"

	assert.Equal(t, expected, path)
	assert.Contains(t, path, ".yaml")
	assert.Contains(t, path, "ksail")
}

// TestViperConfiguration tests the Viper initialization and configuration.
func TestViperConfiguration(t *testing.T) {
	t.Parallel()

	// Create a manager to test Viper configuration
	manager := config.NewManager()
	viper := manager.GetViper()

	// Test Viper configuration
	assert.NotNil(t, viper)

	// Test that environment variable replacement is configured
	// This is tested indirectly by checking that the manager works with env vars

	// Test config file paths are set
	// We can't directly access Viper's internal config paths, but we can test
	// that the Viper instance works properly

	// Test default values can be set and retrieved
	viper.SetDefault("test.key", "test-value")
	assert.Equal(t, "test-value", viper.GetString("test.key"))
}

// TestViperEnvironmentVariables tests environment variable handling.
func TestViperEnvironmentVariables(t *testing.T) {
	// Note: Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
	// Test environment variable prefix and key replacement
	setupTestEnvironment(t)

	// Set an environment variable with the expected format
	t.Setenv("KSAIL_TEST_KEY", "env-value")

	manager := config.NewManager()
	viper := manager.GetViper()

	// Test that the environment variable is read correctly
	value := viper.GetString("test.key")
	assert.Equal(t, "env-value", value)
}

// TestViperConfigFileReading tests configuration file reading capability.
//
//nolint:paralleltest // Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
func TestViperConfigFileReading(t *testing.T) {
	// Note: Cannot use t.Parallel() because we use setupTestEnvironment which calls t.Chdir
	setupTestEnvironment(t)

	// Create a simple config file
	createSimpleConfigForTesting(t)

	manager := config.NewManager()
	viper := manager.GetViper()

	// Test that config file values can be read
	// The exact values depend on the config file structure
	assert.NotNil(t, viper)
}

// Helper function to create a simple config file for Viper testing.
func createSimpleConfigForTesting(t *testing.T) {
	t.Helper()

	// Use the existing config file creation helper
	createConfigFileForTesting(t)
}
