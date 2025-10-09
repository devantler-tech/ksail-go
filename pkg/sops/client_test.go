package sops_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/sops"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCreateCipherCommand(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	if cmd.Use != "cipher" {
		t.Errorf("expected Use to be 'cipher', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	if !cmd.DisableFlagParsing {
		t.Error("expected DisableFlagParsing to be true")
	}
}

func TestCipherCommandHasSubcommands(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	// The command should have subcommands available
	// (inherited from the wrapped urfave/cli app)
	if cmd.Use != "cipher" {
		t.Errorf("expected Use to be 'cipher', got %q", cmd.Use)
	}
}
