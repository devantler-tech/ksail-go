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

func TestDownCmdExecute(t *testing.T) {
	t.Parallel()

	// Test command creation rather than execution since execution requires valid cluster configuration
	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName: "down",
		NewCommand:  cmd.NewDownCmd,
	})
}

func TestDownCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "down",
		NewCommand:  cmd.NewDownCmd,
	})
}
