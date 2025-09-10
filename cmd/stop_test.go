package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStopCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "stop",
		ExpectedUse:    "stop",
		ExpectedShort:  "Stop the Kubernetes cluster",
		NewCommandFunc: cmd.NewStopCmd,
	})
}

func TestStopCmd_Execute(t *testing.T) {
	t.Parallel()

	testCommandExecution(t, cmd.NewStopCmd)
}

func TestStopCmd_Help(t *testing.T) {
	t.Parallel()

	testCommandHelp(t, cmd.NewStopCmd)
}
