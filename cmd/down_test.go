package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewDownCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "down",
		ExpectedUse:    "down",
		ExpectedShort:  "Destroy a cluster",
		NewCommandFunc: cmd.NewDownCmd,
	})
}

func TestDownCmd_Execute(t *testing.T) {
	t.Parallel()

	testCommandExecution(t, cmd.NewDownCmd)
}

func TestDownCmd_Help(t *testing.T) {
	t.Parallel()

	testCommandHelp(t, cmd.NewDownCmd)
}
