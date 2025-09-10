package cmd_test

import (
	"testing"

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
