package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"testing"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

func TestNewInfoCmd(t *testing.T) {
	t.Parallel()

	runtimeContainer := runtime.NewRuntime()
	cmd := NewInfoCmd(runtimeContainer)

	if cmd.Use != "info" {
		t.Fatalf("expected Use to be 'info', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatal("expected Short description to be set")
	}

	// Check that the command has a Run or RunE function (kubectl commands use Run)
	if cmd.Run == nil && cmd.RunE == nil {
		t.Fatal("expected Run or RunE to be set")
	}

	// Check that help can be displayed
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected help to execute without error, got %v", err)
	}

	output := out.String()
	if output == "" {
		t.Fatal("expected help output to be non-empty")
	}

	// Verify the dump subcommand exists
	dumpCmd, _, err := cmd.Find([]string{"dump"})
	if err != nil {
		t.Fatalf("expected to find dump subcommand, got error: %v", err)
	}

	if dumpCmd.Use != "dump" {
		t.Fatalf("expected dump subcommand Use to be 'dump', got %q", dumpCmd.Use)
	}
}
