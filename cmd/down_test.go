package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewDownCmd(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandCreation(t, cmd.SimpleCommandTestData{
		CommandName:   "down",
		NewCommand:    cmd.NewDownCmd,
		ExpectedUse:   "down",
		ExpectedShort: "Destroy a cluster",
	})
}

func TestDownCmd_Execute(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandExecution(t, cmd.SimpleCommandTestData{
		CommandName: "down",
		NewCommand:  cmd.NewDownCmd,
	})
}

func TestDownCmd_Help(t *testing.T) {
	t.Parallel()

	cmd.TestSimpleCommandHelp(t, cmd.SimpleCommandTestData{
		CommandName: "down",
		NewCommand:  cmd.NewDownCmd,
	})
}
