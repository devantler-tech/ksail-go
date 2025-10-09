package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewConnectCmd(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewConnectCmd(runtimeContainer)

	require.NotNil(t, cmd, "expected command to be created")
	require.Equal(t, "connect", cmd.Use, "expected Use to be 'connect'")
	require.Equal(t, "Connect to cluster with k9s", cmd.Short, "expected Short description")
	require.Contains(t, cmd.Long, "k9s terminal UI", "expected Long description to mention k9s")
	require.True(t, cmd.SilenceUsage, "expected SilenceUsage to be true")
	require.NotNil(t, cmd.RunE, "expected RunE to be set")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_LoadsConfig(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")

	// Create a minimal ksail.yaml configuration
	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind
  connection:
    kubeconfig: ` + kubeConfigPath + `
`
	configFile := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err, "failed to write config file")

	// Change to temp directory
	t.Chdir(tempDir)

	// Create a command and config manager
	cmd := &cobra.Command{}

	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout(), selectors...)

	// Load config to verify it works
	err = cfgManager.LoadConfigSilent()
	require.NoError(t, err, "expected config to load successfully")

	cfg := cfgManager.GetConfig()
	require.Equal(t, kubeConfigPath, cfg.Spec.Connection.Kubeconfig,
		"expected kubeconfig path to be loaded from config")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_UsesDefaultKubeconfig(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Create a minimal ksail.yaml configuration without kubeConfigPath
	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind
`
	configFile := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err, "failed to write config file")

	// Change to temp directory
	t.Chdir(tempDir)

	// Create a command and config manager
	cmd := &cobra.Command{}

	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout(), selectors...)

	// Load config to verify it works
	err = cfgManager.LoadConfigSilent()
	require.NoError(t, err, "expected config to load successfully")

	cfg := cfgManager.GetConfig()
	require.Empty(t, cfg.Spec.Connection.Kubeconfig,
		"expected kubeconfig path to be empty, will use default")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_ConfigLoadError(t *testing.T) {
	// Create a temporary directory with invalid config
	tempDir := t.TempDir()

	// Create an invalid configuration file
	configContent := `invalid yaml content [[[`
	configFile := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err, "failed to write config file")

	// Change to temp directory
	t.Chdir(tempDir)

	// Create a command and config manager
	cmd := &cobra.Command{}

	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout(), selectors...)

	// Run should fail to load config
	err = HandleConnectRunE(cmd, cfgManager, []string{})
	require.Error(t, err, "expected error when config is invalid")
	require.Contains(
		t,
		err.Error(),
		"load configuration",
		"expected error to indicate config loading failure",
	)
}
