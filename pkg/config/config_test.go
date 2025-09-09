package config_test

import (
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear any existing environment variables that might affect the test
	envVarsToClean := []string{
		"KSAIL_DISTRIBUTION",
		"KSAIL_ALL", 
		"KSAIL_METADATA_NAME",
		"KSAIL_SPEC_CONNECTION_KUBECONFIG",
	}
	for _, envVar := range envVarsToClean {
		if originalValue := os.Getenv(envVar); originalValue != "" {
			_ = os.Unsetenv(envVar)
			defer func(envVar, originalValue string) {
				_ = os.Setenv(envVar, originalValue)
			}(envVar, originalValue)
		}
	}

	// Load config without any files or env vars
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	// Test defaults - distribution should be empty from CLI/env, but cluster should have defaults
	assert.Equal(t, "", cfg.Distribution) // No CLI default anymore
	assert.False(t, cfg.All)
	assert.Equal(t, "ksail-default", cfg.Cluster.Name)
	assert.Equal(t, "kind.yaml", cfg.Cluster.DistributionConfig)
	assert.Equal(t, "k8s", cfg.Cluster.SourceDirectory)
	assert.Equal(t, "~/.kube/config", cfg.Cluster.Connection.Kubeconfig)
	assert.Equal(t, "kind-ksail-default", cfg.Cluster.Connection.Context)
	assert.Equal(t, "5m", cfg.Cluster.Connection.Timeout)
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Set environment variables - using the correct hierarchical structure
	envVars := map[string]string{
		"KSAIL_DISTRIBUTION":                        "K3d",
		"KSAIL_ALL":                                "true",
		"KSAIL_METADATA_NAME":                      "test-cluster",
		"KSAIL_SPEC_CONNECTION_KUBECONFIG":         "/custom/kubeconfig",
	}

	// Set env vars and defer cleanup
	for key, value := range envVars {
		_ = os.Setenv(key, value)
		defer func(key string) {
			_ = os.Unsetenv(key)
		}(key)
	}

	// Load config
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	// Test environment variable overrides
	assert.Equal(t, "K3d", cfg.Distribution)
	assert.True(t, cfg.All)
	assert.Equal(t, "test-cluster", cfg.Cluster.Name)
	assert.Equal(t, "/custom/kubeconfig", cfg.Cluster.Connection.Kubeconfig)
}

func TestInitializeViper(t *testing.T) {
	// Clear any existing environment variables that might affect the test
	envVarsToClean := []string{
		"KSAIL_DISTRIBUTION",
		"KSAIL_ALL", 
		"KSAIL_METADATA_NAME",
		"KSAIL_SPEC_CONNECTION_KUBECONFIG",
	}
	for _, envVar := range envVarsToClean {
		if originalValue := os.Getenv(envVar); originalValue != "" {
			_ = os.Unsetenv(envVar)
			defer func(envVar, originalValue string) {
				_ = os.Setenv(envVar, originalValue)
			}(envVar, originalValue)
		}
	}

	v := config.InitializeViper()
	assert.NotNil(t, v)

	// Test that no CLI defaults are set (following Viper best practices)
	assert.Equal(t, "", v.GetString("distribution"))
	assert.False(t, v.GetBool("all"))
	assert.Equal(t, "", v.GetString("metadata.name")) // No defaults in Viper anymore
}

func TestGetConfigFilePath(t *testing.T) {
	path := config.GetConfigFilePath()
	assert.Equal(t, "ksail.yaml", path)
}