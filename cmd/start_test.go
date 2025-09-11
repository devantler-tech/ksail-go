package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStartCmd(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandCreation(t, cmd.SimpleCommandTestData{
		CommandName:   "start",
		NewCommand:    cmd.NewStartCmd,
		ExpectedUse:   "start",
		ExpectedShort: "Start a stopped cluster",
	})
}

func TestStartCmd_Execute(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandExecution(t, cmd.SimpleCommandTestData{
		CommandName: "start",
		NewCommand:  cmd.NewStartCmd,
	})
}

func TestStartCmd_Help(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandHelp(t, cmd.SimpleCommandTestData{
		CommandName: "start",
		NewCommand:  cmd.NewStartCmd,
	})
}
