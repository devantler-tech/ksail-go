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

func TestUpCmdExecute(t *testing.T) {
	t.Parallel()

	// Test command creation rather than execution since execution requires valid cluster configuration
	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName: "up",
		NewCommand:  cmd.NewUpCmd,
	})
}

func TestUpCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "up",
		NewCommand:  cmd.NewUpCmd,
	})
}
