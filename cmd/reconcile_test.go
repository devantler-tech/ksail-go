package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/internal/cmd/testutils"
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
