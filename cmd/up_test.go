package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
)

func TestNewUpCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "up",
		NewCommand:    cmd.NewUpCmd,
		ExpectedUse:   "up",
		ExpectedShort: "Start the Kubernetes cluster",
	})
}

func TestUpCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "up",
		NewCommand:  cmd.NewUpCmd,
	})
}

func TestUpCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "up",
		NewCommand:  cmd.NewUpCmd,
	})
}
