package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStatusCmd(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandCreation(t, cmd.SimpleCommandTestData{
		CommandName:   "status",
		NewCommand:    cmd.NewStatusCmd,
		ExpectedUse:   "status",
		ExpectedShort: "Show status of the Kubernetes cluster",
	})
}

func TestStatusCmd_Execute(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandExecution(t, cmd.SimpleCommandTestData{
		CommandName: "status",
		NewCommand:  cmd.NewStatusCmd,
	})
}

func TestStatusCmd_Help(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandHelp(t, cmd.SimpleCommandTestData{
		CommandName: "status",
		NewCommand:  cmd.NewStatusCmd,
	})
}
