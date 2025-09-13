package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/internal/cmd/testutils"
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

func TestStartCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "start",
		NewCommand:  cmd.NewStartCmd,
	})
}

func TestStartCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "start",
		NewCommand:  cmd.NewStartCmd,
	})
}
