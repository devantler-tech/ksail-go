package cmd_test

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

// CommandTestConfig holds the configuration for testing a command.
type CommandTestConfig struct {
	CommandName    string
	ExpectedUse    string
	ExpectedShort  string
	NewCommandFunc func() *cobra.Command
}

// testCommandCreation tests the basic creation and properties of a command.
func testCommandCreation(t *testing.T, cfg CommandTestConfig) {
	t.Helper()

	cmd := cfg.NewCommandFunc()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != cfg.ExpectedUse {
		t.Fatalf("expected Use to be %q, got %q", cfg.ExpectedUse, cmd.Use)
	}

	if cmd.Short != cfg.ExpectedShort {
		t.Fatalf("expected Short description to be %q, got %q", cfg.ExpectedShort, cmd.Short)
	}
}

// testCommandExecution tests the execution of a command.
func testCommandExecution(t *testing.T, newCommandFunc func() *cobra.Command) {
	t.Helper()

	var out bytes.Buffer

	cmd := newCommandFunc()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

// testCommandHelp tests the help output of a command.
func testCommandHelp(t *testing.T, newCommandFunc func() *cobra.Command) {
	t.Helper()

	var out bytes.Buffer

	cmd := newCommandFunc()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}
