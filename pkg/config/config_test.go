package config_test

import (
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_LoadCluster_Defaults(t *testing.T) {
	// Clear any existing environment variables that might affect the test
	envVarsToClean := []string{
		"KSAIL_SPEC_DISTRIBUTION",
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

	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tempDir)

	// Load cluster without any files or env vars
	manager := config.NewManager()
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Test defaults
	assert.Equal(t, "ksail-default", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionKind, cluster.Spec.Distribution)
	assert.Equal(t, "kind.yaml", cluster.Spec.DistributionConfig)
	assert.Equal(t, "k8s", cluster.Spec.SourceDirectory)
	assert.Equal(t, "~/.kube/config", cluster.Spec.Connection.Kubeconfig)
	assert.Equal(t, "kind-ksail-default", cluster.Spec.Connection.Context)
}

func TestManager_LoadCluster_EnvironmentVariables(t *testing.T) {
	// Set environment variables - using the correct hierarchical structure
	_ = os.Setenv("KSAIL_METADATA_NAME", "test-cluster")
	_ = os.Setenv("KSAIL_SPEC_DISTRIBUTION", "K3d")
	_ = os.Setenv("KSAIL_SPEC_CONNECTION_KUBECONFIG", "/custom/path/kubeconfig")
	defer func() {
		_ = os.Unsetenv("KSAIL_METADATA_NAME")
		_ = os.Unsetenv("KSAIL_SPEC_DISTRIBUTION")
		_ = os.Unsetenv("KSAIL_SPEC_CONNECTION_KUBECONFIG")
	}()

	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldDir) }()
	_ = os.Chdir(tempDir)

	manager := config.NewManager()
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Environment variables should override defaults
	assert.Equal(t, "test-cluster", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)
	assert.Equal(t, "/custom/path/kubeconfig", cluster.Spec.Connection.Kubeconfig)
}

func TestNewManager(t *testing.T) {
	manager := config.NewManager()
	require.NotNil(t, manager)
	require.NotNil(t, manager.GetViper())
}
