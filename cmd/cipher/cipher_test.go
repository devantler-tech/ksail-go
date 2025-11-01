package cipher_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/cipher"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

func TestNewCipherCmd(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cmd := cipher.NewCipherCmd(rt)

	if cmd == nil {
		t.Fatal("expected non-nil command")
	}

	if cmd.Use != "cipher" {
		t.Errorf("expected Use to be 'cipher', got %q", cmd.Use)
	}

	// Verify the short description is set
	if cmd.Short == "" {
		t.Error("expected Short description to be set")
	}

	// Verify subcommands are registered
	if !cmd.HasSubCommands() {
		t.Error("expected cipher command to have subcommands")
	}

	// Verify encrypt subcommand exists
	encryptCmd := findSubcommand(cmd, "encrypt")
	if encryptCmd == nil {
		t.Error("expected encrypt subcommand to be registered")
	}

	// Verify decrypt subcommand exists
	decryptCmd := findSubcommand(cmd, "decrypt")
	if decryptCmd == nil {
		t.Error("expected decrypt subcommand to be registered")
	}
}

func TestCipherCommandHelp(t *testing.T) {
	t.Parallel()

	rt := runtime.NewRuntime()
	cmd := cipher.NewCipherCmd(rt)

	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing help: %v", err)
	}

	output := out.String()
	if output == "" {
		t.Error("expected non-empty help output")
	}

	// Verify help mentions subcommands
	if !contains(output, "encrypt") {
		t.Error("expected help to mention encrypt subcommand")
	}
	if !contains(output, "decrypt") {
		t.Error("expected help to mention decrypt subcommand")
	}
}

// Helper function to find a subcommand by name
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, sub := range cmd.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
