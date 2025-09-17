package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
)

func TestNewStartCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "start",
		NewCommand:    cmd.NewStartCmd,
		ExpectedUse:   "start",
		ExpectedShort: "Start a stopped cluster",
	})
}

func TestStartCmdExecute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "start",
		NewCommand:  cmd.NewStartCmd,
	})
}

func TestStartCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "start",
		NewCommand:  cmd.NewStartCmd,
	})
}
