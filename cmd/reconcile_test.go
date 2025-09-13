package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReconcileCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "reconcile",
		NewCommand:    cmd.NewReconcileCmd,
		ExpectedUse:   "reconcile",
		ExpectedShort: "Reconcile workloads in the cluster",
	})
}

func TestReconcileCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
}

func TestReconcileCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
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

	mockManager := config.NewMockConfigManager(t)
	mockManager.EXPECT().LoadCluster().Return(nil, testutils.ErrTestConfigLoadError)

	err := cmd.HandleReconcileRunE(testCmd, mockManager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}
