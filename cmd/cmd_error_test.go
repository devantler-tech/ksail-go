package cmd_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors to comply with err113.
var errTestConfigLoadError = errors.New("test config load error")

// TestHandleInitRunE_Success tests successful init command execution.
func TestHandleInitRunE_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()

	err := cmd.HandleInitRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ project initialized successfully")
	assert.Contains(t, out.String(), "► Distribution:")
	assert.Contains(t, out.String(), "► Source directory:")
}

// TestHandleInitRunE_Error tests init command with config load error.
func TestHandleInitRunE_Error(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()
	manager.SetTestErrorHook(errTestConfigLoadError)

	err := cmd.HandleInitRunE(testCmd, manager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}

// TestHandleListRunE_Success tests successful list command execution.
func TestHandleListRunE_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	// Add the --all flag to the command like the real command would have
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	manager := config.NewManager()

	err := cmd.HandleListRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Listing running clusters (stub implementation)")
	assert.Contains(t, out.String(), "► Distribution filter:")
}

// TestHandleListRunE_AllFlag tests list command with --all flag.
func TestHandleListRunE_AllFlag(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")
	// Set the flag value
	err := testCmd.Flags().Set("all", "true")
	require.NoError(t, err)

	manager := config.NewManager()

	err = cmd.HandleListRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Listing all clusters (stub implementation)")
	assert.Contains(t, out.String(), "► Distribution filter:")
}

// TestHandleListRunE_Error tests list command with config load error.
func TestHandleListRunE_Error(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	manager := config.NewManager()
	manager.SetTestErrorHook(errTestConfigLoadError)

	err := cmd.HandleListRunE(testCmd, manager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}

// TestHandleReconcileRunE_Success tests successful reconcile command execution.
func TestHandleReconcileRunE_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()

	err := cmd.HandleReconcileRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Workloads reconciled successfully (stub implementation)")
	assert.Contains(t, out.String(), "► Reconciliation tool:")
	assert.Contains(t, out.String(), "► Source directory:")
	assert.Contains(t, out.String(), "► Context:")
}

// TestHandleReconcileRunE_Error tests reconcile command with config load error.
func TestHandleReconcileRunE_Error(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()
	manager.SetTestErrorHook(errTestConfigLoadError)

	err := cmd.HandleReconcileRunE(testCmd, manager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}

// TestHandleStatusRunE_Success tests successful status command execution.
func TestHandleStatusRunE_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()

	err := cmd.HandleStatusRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Cluster status: Running (stub implementation)")
	assert.Contains(t, out.String(), "► Context:")
	assert.Contains(t, out.String(), "► Kubeconfig:")
}

// TestHandleStatusRunE_Error tests status command with config load error.
func TestHandleStatusRunE_Error(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()
	manager.SetTestErrorHook(errTestConfigLoadError)

	err := cmd.HandleStatusRunE(testCmd, manager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}
