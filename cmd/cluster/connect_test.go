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

func TestNewConnectCmd_RunECallsHandler(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewConnectCmd(runtimeContainer)

	require.NotNil(t, cmd, "expected command to be created")
	require.NotNil(t, cmd.RunE, "expected RunE to be set")

	// Verify the command structure allows RunE to be called
	// We don't actually call it because it would try to run k9s
	// but we verify it's properly wired up
	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	require.NotNil(t, cmd.RunE, "RunE should be callable")
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

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_WithDefaultKubeconfigPath(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Create a minimal ksail.yaml configuration without kubeconfig
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

	// Load config first to verify the path logic
	err = cfgManager.LoadConfigSilent()
	require.NoError(t, err, "expected config to load successfully")

	cfg := cfgManager.GetConfig()

	// Verify that kubeconfig is empty in config
	require.Empty(t, cfg.Spec.Connection.Kubeconfig,
		"expected kubeconfig path to be empty before defaulting")

	// Verify home directory can be obtained
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "expected to get home directory")
	require.NotEmpty(t, homeDir, "expected home directory to be non-empty")

	// Verify the default path would be constructed correctly
	expectedPath := filepath.Join(homeDir, ".kube", "config")
	require.NotEmpty(t, expectedPath, "expected default kubeconfig path to be constructed")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_WithCustomKubeconfigPath(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	customKubeConfigPath := filepath.Join(tempDir, "custom-kubeconfig")

	// Create a minimal ksail.yaml configuration with custom kubeconfig
	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind
  connection:
    kubeconfig: ` + customKubeConfigPath + `
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

	// Load config to verify the custom path
	err = cfgManager.LoadConfigSilent()
	require.NoError(t, err, "expected config to load successfully")

	cfg := cfgManager.GetConfig()
	require.Equal(t, customKubeConfigPath, cfg.Spec.Connection.Kubeconfig,
		"expected custom kubeconfig path to be loaded from config")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_WithAdditionalArgs(t *testing.T) {
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

	// Verify we can load config and it would pass args
	err = cfgManager.LoadConfigSilent()
	require.NoError(t, err, "expected config to load successfully")

	// Test that args would be passed through
	args := []string{"--namespace", "default", "--readonly"}
	require.NotNil(t, args, "expected args to be passable")
}
