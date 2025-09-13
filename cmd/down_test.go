package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
)

func TestNewDownCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "down",
		NewCommand:    cmd.NewDownCmd,
		ExpectedUse:   "down",
		ExpectedShort: "Destroy a cluster",
	})
}

func TestDownCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "down",
		NewCommand:  cmd.NewDownCmd,
	})
}

func TestDownCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "down",
		NewCommand:  cmd.NewDownCmd,
	})
}
