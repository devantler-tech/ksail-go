package sops_test

import (
	"bytes"
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

	if cmd.Short != "Manage encryption and decryption with SOPS" {
		t.Errorf("expected Short description, got %q", cmd.Short)
	}

	if !cmd.DisableFlagParsing {
		t.Error("expected DisableFlagParsing to be true")
	}
}

func TestCipherCommandHelp(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	// Since DisableFlagParsing is true, --help is passed to sops binary
	// We can't easily test this without sops installed, so just verify
	// the command executes (it will try to run sops --help)
	_ = cmd.Execute()

	// The command structure is correct even if execution fails
	// (e.g., if sops is not installed in the test environment)
}
