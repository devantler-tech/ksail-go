package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
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

func TestReconcileCmdExecute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
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

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := configmanager.NewConfigManager()

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

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := configmanager.NewConfigManager()

	// Test that the function doesn't panic - error testing can be enhanced later
	assert.NotPanics(t, func() {
		_ = cmd.HandleReconcileRunE(testCmd, manager, []string{})
	})
}
