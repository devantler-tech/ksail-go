package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewReconcileCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "reconcile",
		ExpectedUse:    "reconcile",
		ExpectedShort:  "Reconcile workloads in the cluster",
		NewCommandFunc: cmd.NewReconcileCmd,
	})
}

func TestReconcileCmd_Execute(t *testing.T) {
	t.Parallel()

	testCommandExecution(t, cmd.NewReconcileCmd)
}

func TestReconcileCmd_Help(t *testing.T) {
	t.Parallel()

	testCommandHelp(t, cmd.NewReconcileCmd)
}
