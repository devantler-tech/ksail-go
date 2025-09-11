package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewReconcileCmd(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandCreation(t, cmd.SimpleCommandTestData{
		CommandName:   "reconcile",
		NewCommand:    cmd.NewReconcileCmd,
		ExpectedUse:   "reconcile",
		ExpectedShort: "Reconcile workloads in the cluster",
	})
}

func TestReconcileCmd_Execute(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandExecution(t, cmd.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
}

func TestReconcileCmd_Help(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandHelp(t, cmd.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
}
