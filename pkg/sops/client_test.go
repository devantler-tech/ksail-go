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

func TestCipherCommandHasLongDescription(t *testing.T) {
	t.Parallel()

	client := sops.NewClient()
	cmd := client.CreateCipherCommand()

	// The command should have a long description that mentions the sops dependency
	if cmd.Long == "" {
		t.Error("expected Long description to be set")
	}

	// Verify the description mentions the sops binary dependency
	if !contains(cmd.Long, "sops") {
		t.Error("expected Long description to mention 'sops' binary")
	}

	if !contains(cmd.Long, "Dependencies") {
		t.Error("expected Long description to mention Dependencies")
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr || containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}
