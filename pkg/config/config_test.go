package config_test

import (
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Parallel()

	// Clear any existing environment variables that might affect the test
	envVarsToClean := []string{
		"KSAIL_DISTRIBUTION",
		"KSAIL_ALL", 
		"KSAIL_CLUSTER_NAME",
		"KSAIL_CLUSTER_CONNECTION_KUBECONFIG",
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

	// Test defaults
	assert.Equal(t, "Kind", cfg.Distribution)
	assert.False(t, cfg.All)
	assert.Equal(t, "ksail-default", cfg.Cluster.Name)
	assert.Equal(t, "kind.yaml", cfg.Cluster.DistributionConfig)
	assert.Equal(t, "k8s", cfg.Cluster.SourceDirectory)
	assert.Equal(t, "~/.kube/config", cfg.Cluster.Connection.Kubeconfig)
	assert.Equal(t, "kind-ksail-default", cfg.Cluster.Connection.Context)
	assert.Equal(t, "5m", cfg.Cluster.Connection.Timeout)
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	t.Parallel()

	// Set environment variables
	envVars := map[string]string{
		"KSAIL_DISTRIBUTION":                   "K3d",
		"KSAIL_ALL":                           "true",
		"KSAIL_CLUSTER_NAME":                  "test-cluster",
		"KSAIL_CLUSTER_CONNECTION_KUBECONFIG": "/custom/kubeconfig",
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
	t.Parallel()

	v := config.InitializeViper()
	assert.NotNil(t, v)

	// Test that defaults are set
	assert.Equal(t, "Kind", v.GetString("distribution"))
	assert.False(t, v.GetBool("all"))
	assert.Equal(t, "ksail-default", v.GetString("cluster.name"))
}

func TestGetConfigFilePath(t *testing.T) {
	t.Parallel()

	path := config.GetConfigFilePath()
	assert.Equal(t, "ksail.yaml", path)
}