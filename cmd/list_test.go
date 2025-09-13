package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/internal/cmd/testutils"
	"github.com/spf13/cobra"
)

func TestNewListCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewListCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "list" {
		t.Fatalf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short != "List clusters" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestListCmd_Execute_Default(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "list",
		NewCommand:  cmd.NewListCmd,
	})
}

func TestListCmd_Execute_All(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "list",
		NewCommand: func() *cobra.Command {
			cmd := cmd.NewListCmd()
			cmd.SetArgs([]string{"--all"})

			return cmd
		},
	})
}

func TestListCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		NewCommand: cmd.NewListCmd,
	})
}

func TestListCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewListCmd()

	// Act & Assert
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("expected all flag to exist")
	}

	if allFlag.DefValue != "false" {
		t.Fatalf("expected all flag default to be 'false', got %q", allFlag.DefValue)
	}
}
