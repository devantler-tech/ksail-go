package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
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

//nolint:paralleltest // Cannot use t.Parallel() with t.Chdir() in Go 1.25.1+
func TestReconcileCmdExecute(t *testing.T) {
	testutils.TestCmdExecuteInCleanDir(t, cmd.NewReconcileCmd, "reconcile")
}

func TestReconcileCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
}

// TestHandleReconcileRunE_Success tests successful reconcile command execution.
func TestHandleReconcileRunESuccess(t *testing.T) {
	t.Parallel()

	testCmd, out := testutils.SetupCommandWithOutput()
	manager := testutils.CreateDefaultConfigManager()

	err := cmd.HandleReconcileRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Workloads reconciled successfully (stub implementation)")
	assert.Contains(t, out.String(), "► Reconciliation tool:")
	assert.Contains(t, out.String(), "► Source directory:")
	assert.Contains(t, out.String(), "► Context:")
}

// TestHandleReconcileRunE_Error tests reconcile command with config load error.
func TestHandleReconcileRunEError(t *testing.T) {
	t.Parallel()

	testCmd, _ := testutils.SetupCommandWithOutput()
	manager := configmanager.NewConfigManager()

	// Test that the function doesn't panic - error testing can be enhanced later
	assert.NotPanics(t, func() {
		_ = cmd.HandleReconcileRunE(testCmd, manager, []string{})
	})
}
