package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
)

func TestNewStatusCmd(t *testing.T) {
	t.Parallel()

	testCommandCreation(t, CommandTestConfig{
		CommandName:    "status",
		ExpectedUse:    "status",
		ExpectedShort:  "Show status of the Kubernetes cluster",
		NewCommandFunc: cmd.NewStatusCmd,
	})
}

func TestStatusCmd_Execute(t *testing.T) {
	t.Parallel()

	testCommandExecution(t, cmd.NewStatusCmd)
}

func TestStatusCmd_Help(t *testing.T) {
	t.Parallel()

	testCommandHelp(t, cmd.NewStatusCmd)
}
