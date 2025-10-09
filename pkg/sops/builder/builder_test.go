package builder_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/sops/builder"
)

func TestNewSopsApp(t *testing.T) {
	t.Parallel()

	app := builder.NewSopsApp()

	if app == nil {
		t.Fatal("expected non-nil app")
	}

	if app.Name != "cipher" {
		t.Errorf("expected app name to be 'cipher', got %q", app.Name)
	}

	if app.Usage == "" {
		t.Error("expected app usage to be set")
	}

	// Check that commands are defined
	expectedCommands := []string{
		"encrypt", "decrypt", "rotate", "edit",
		"set", "unset", "updatekeys", "groups",
	}
	if len(app.Commands) != len(expectedCommands) {
		t.Errorf("expected %d commands, got %d", len(expectedCommands), len(app.Commands))
	}

	// Verify each command exists
	for _, expectedCmd := range expectedCommands {
		found := false

		for _, cmd := range app.Commands {
			if cmd.Name == expectedCmd {
				found = true

				break
			}
		}

		if !found {
			t.Errorf("expected command %q not found", expectedCmd)
		}
	}
}

func TestSopsAppCommands(t *testing.T) {
	t.Parallel()

	app := builder.NewSopsApp()

	// Verify each command has the required properties
	for _, cmd := range app.Commands {
		if cmd.Name == "" {
			t.Error("command name should not be empty")
		}

		if cmd.Usage == "" {
			t.Errorf("command %q should have usage text", cmd.Name)
		}

		if cmd.Action == nil {
			t.Errorf("command %q should have an action", cmd.Name)
		}

		if !cmd.SkipFlagParsing {
			t.Errorf("command %q should have SkipFlagParsing set to true", cmd.Name)
		}
	}
}
