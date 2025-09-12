package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStopCmd(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandCreation(t, cmd.SimpleCommandTestData{
		CommandName:   "stop",
		NewCommand:    cmd.NewStopCmd,
		ExpectedUse:   "stop",
		ExpectedShort: "Stop the Kubernetes cluster",
	})
}

func TestStopCmd_Execute(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandExecution(t, cmd.SimpleCommandTestData{
		CommandName: "stop",
		NewCommand:  cmd.NewStopCmd,
	})
}

func TestStopCmd_Help(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandHelp(t, cmd.SimpleCommandTestData{
		CommandName: "stop",
		NewCommand:  cmd.NewStopCmd,
	})
}
