package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/internal/cmd/testutils"
)

func TestNewStatusCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "status",
		NewCommand:    cmd.NewStatusCmd,
		ExpectedUse:   "status",
		ExpectedShort: "Show status of the Kubernetes cluster",
	})
}

func TestStatusCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "status",
		NewCommand:  cmd.NewStatusCmd,
	})
}

func TestStatusCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "status",
		NewCommand:  cmd.NewStatusCmd,
	})
}
