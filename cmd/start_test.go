package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStartCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "start",
		ExpectedUse:    "start",
		ExpectedShort:  "Start a stopped cluster",
		NewCommandFunc: cmd.NewStartCmd,
	})
}

func TestStartCmd_Execute(t *testing.T) {
	t.Parallel()

	testCommandExecution(t, cmd.NewStartCmd)
}

func TestStartCmd_Help(t *testing.T) {
	t.Parallel()

	testCommandHelp(t, cmd.NewStartCmd)
}
