// Package testutils provides testing helpers for command testing.
package testutils

import (
	"bytes"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

// SimpleCommandTestData holds test data for simple command testing.
type SimpleCommandTestData struct {
	CommandName   string
	NewCommand    func() *cobra.Command
	ExpectedUse   string
	ExpectedShort string
}

// TestSimpleCommandCreation tests command creation with common pattern.
func TestSimpleCommandCreation(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	cmd := data.NewCommand()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != data.ExpectedUse {
		t.Fatalf("expected Use to be %q, got %q", data.ExpectedUse, cmd.Use)
	}

	if cmd.Short != data.ExpectedShort {
		t.Fatalf("expected Short description to be %q, got %q", data.ExpectedShort, cmd.Short)
	}
}

// TestSimpleCommandExecution tests command execution with common pattern.
func TestSimpleCommandExecution(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	var out bytes.Buffer

	cmd := data.NewCommand()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

// TestSimpleCommandHelp tests command help output with common pattern.
func TestSimpleCommandHelp(t *testing.T, data SimpleCommandTestData) {
	t.Helper()

	var out bytes.Buffer

	cmd := data.NewCommand()
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}
