package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
)

func TestNewStopCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "stop",
		NewCommand:    cmd.NewStopCmd,
		ExpectedUse:   "stop",
		ExpectedShort: "Stop the Kubernetes cluster",
	})
}

func TestStopCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "stop",
		NewCommand:  cmd.NewStopCmd,
	})
}

func TestStopCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "stop",
		NewCommand:  cmd.NewStopCmd,
	})
}
