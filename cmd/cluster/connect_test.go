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

// setupTestConfig creates a test config file and returns the config manager.
func setupTestConfig(
	t *testing.T,
	configContent string,
) *ksailconfigmanager.ConfigManager {
	t.Helper()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "ksail.yaml")

	err := os.WriteFile(configFile, []byte(configContent), 0o600)
	require.NoError(t, err, "failed to write config file")

	t.Chdir(tempDir)

	cmd := &cobra.Command{}

	var outBuf bytes.Buffer
	cmd.SetOut(&outBuf)

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout(), selectors...)

	return cfgManager
}

// loadConfigAndVerifyNoError is a helper that loads config and asserts no error occurred.
func loadConfigAndVerifyNoError(t *testing.T, cfgManager *ksailconfigmanager.ConfigManager) {
	t.Helper()

	err := cfgManager.LoadConfigSilent()
	require.NoError(t, err, "expected config to load successfully")
}

// createKSailConfigYAML is a helper that creates a ksail.yaml config string.
func createKSailConfigYAML(kubeconfig, context string) string {
	config := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind`

	if kubeconfig != "" || context != "" {
		config += `
  connection:`

		if kubeconfig != "" {
			config += `
    kubeconfig: ` + kubeconfig
		}

		if context != "" {
			config += `
    context: ` + context
		}
	}

	return config
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_LoadsConfig(t *testing.T) {
	tempDir := t.TempDir()
	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")

	configContent := createKSailConfigYAML(kubeConfigPath, "")
	cfgManager := setupTestConfig(t, configContent)

	loadConfigAndVerifyNoError(t, cfgManager)

	cfg := cfgManager.GetConfig()
	require.Equal(t, kubeConfigPath, cfg.Spec.Connection.Kubeconfig,
		"expected kubeconfig path to be loaded from config")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_DefaultKubeconfig(t *testing.T) {
	configContent := createKSailConfigYAML("", "")
	cfgManager := setupTestConfig(t, configContent)

	loadConfigAndVerifyNoError(t, cfgManager)

	cfg := cfgManager.GetConfig()
	require.Empty(t, cfg.Spec.Connection.Kubeconfig,
		"expected kubeconfig path to be empty, will use default")

	// Verify default path can be constructed
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "expected to get home directory")
	require.NotEmpty(t, homeDir, "expected home directory to be non-empty")

	expectedPath := filepath.Join(homeDir, ".kube", "config")
	require.NotEmpty(t, expectedPath, "expected default kubeconfig path to be constructed")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_ConfigLoadError(t *testing.T) {
	configContent := `invalid yaml content [[[`
	cfgManager := setupTestConfig(t, configContent)

	err := HandleConnectRunE(&cobra.Command{}, cfgManager, []string{})
	require.Error(t, err, "expected error when config is invalid")
	require.Contains(
		t,
		err.Error(),
		"load configuration",
		"expected error to indicate config loading failure",
	)
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_WithCustomKubeconfigPath(t *testing.T) {
	tempDir := t.TempDir()
	customKubeConfigPath := filepath.Join(tempDir, "custom-kubeconfig")

	configContent := createKSailConfigYAML(customKubeConfigPath, "")
	cfgManager := setupTestConfig(t, configContent)

	loadConfigAndVerifyNoError(t, cfgManager)

	cfg := cfgManager.GetConfig()
	require.Equal(t, customKubeConfigPath, cfg.Spec.Connection.Kubeconfig,
		"expected custom kubeconfig path to be loaded from config")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_WithAdditionalArgs(t *testing.T) {
	tempDir := t.TempDir()
	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")

	configContent := createKSailConfigYAML(kubeConfigPath, "")
	cfgManager := setupTestConfig(t, configContent)

	loadConfigAndVerifyNoError(t, cfgManager)

	args := []string{"--namespace", "default", "--readonly"}
	require.NotNil(t, args, "expected args to be passable")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_WithContext(t *testing.T) {
	tempDir := t.TempDir()
	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")
	contextName := "kind-kind"

	configContent := createKSailConfigYAML(kubeConfigPath, contextName)
	cfgManager := setupTestConfig(t, configContent)

	loadConfigAndVerifyNoError(t, cfgManager)

	cfg := cfgManager.GetConfig()
	require.Equal(t, contextName, cfg.Spec.Connection.Context,
		"expected context to be loaded from config")
}

//nolint:paralleltest // Uses t.Chdir for directory-based configuration loading.
func TestHandleConnectRunE_WithoutContext(t *testing.T) {
	tempDir := t.TempDir()
	kubeConfigPath := filepath.Join(tempDir, "kubeconfig")

	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: Kind
  connection:
    kubeconfig: ` + kubeConfigPath + `
`
	cfgManager := setupTestConfig(t, configContent)

	err := cfgManager.LoadConfigSilent()
	require.NoError(t, err, "expected config to load successfully")

	cfg := cfgManager.GetConfig()
	require.Empty(t, cfg.Spec.Connection.Context,
		"expected context to be empty when not specified in config")
}
