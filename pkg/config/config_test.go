package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestManager_LoadCluster_ConfigFile(t *testing.T) {
	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()

	defer func() { _ = os.Chdir(oldDir) }()

	_ = os.Chdir(tempDir)

	// Create a ksail.yaml config file
	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: config-test-cluster
spec:
  distribution: K3d
  sourceDirectory: config-k8s
  connection:
    kubeconfig: /config/path/kubeconfig
    context: config-context
    timeout: 60s
`
	err := os.WriteFile("ksail.yaml", []byte(configContent), 0o600)
	require.NoError(t, err)

	manager := config.NewManager()
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Test config file values are loaded
	assert.Equal(t, "config-test-cluster", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)
	assert.Equal(t, "config-k8s", cluster.Spec.SourceDirectory)
	assert.Equal(t, "/config/path/kubeconfig", cluster.Spec.Connection.Kubeconfig)
	assert.Equal(t, "config-context", cluster.Spec.Connection.Context)
	assert.Equal(t, metav1.Duration{Duration: 60 * time.Second}, cluster.Spec.Connection.Timeout)
}

func TestManager_LoadCluster_MixedConfiguration(t *testing.T) {
	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()

	defer func() { _ = os.Chdir(oldDir) }()

	_ = os.Chdir(tempDir)

	// Create a ksail.yaml config file with some values
	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: config-cluster
spec:
  distribution: K3d
  sourceDirectory: config-k8s
  connection:
    kubeconfig: /config/path/kubeconfig
    context: config-context
`
	err := os.WriteFile("ksail.yaml", []byte(configContent), 0o600)
	require.NoError(t, err)

	// Set environment variables (should override config file)
	_ = os.Setenv("KSAIL_METADATA_NAME", "env-cluster")
	_ = os.Setenv("KSAIL_SPEC_CONNECTION_KUBECONFIG", "/env/path/kubeconfig")

	defer func() {
		_ = os.Unsetenv("KSAIL_METADATA_NAME")
		_ = os.Unsetenv("KSAIL_SPEC_CONNECTION_KUBECONFIG")
	}()

	manager := config.NewManager()
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Test precedence: env vars override config file
	assert.Equal(t, "env-cluster", cluster.Metadata.Name)                       // From env var
	assert.Equal(t, "/env/path/kubeconfig", cluster.Spec.Connection.Kubeconfig) // From env var
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)        // From config file
	assert.Equal(t, "config-k8s", cluster.Spec.SourceDirectory)                 // From config file
	assert.Equal(t, "config-context", cluster.Spec.Connection.Context)          // From config file
}
