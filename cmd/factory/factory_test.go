package factory_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/spf13/cobra"
)

func TestNewCobraCommand(t *testing.T) {
	t.Parallel()

	// Arrange
	use := "test"
	short := "Test command"
	long := "This is a test command for testing purposes"
	runE := func(_ *cobra.Command, _ []string) error { return nil }

	// Act
	cmd := factory.NewCobraCommand(use, short, long, runE)

	// Assert
	assertBasicCommandProperties(t, cmd, use, short, long)

	if cmd.SilenceErrors != false {
		t.Error("expected SilenceErrors to be false for regular command")
	}

	if cmd.SilenceUsage != false {
		t.Error("expected SilenceUsage to be false for regular command")
	}
}

func TestNewCobraCommandWithFlags(t *testing.T) {
	t.Parallel()

	// Arrange
	use := "flagged"
	short := "Flagged command"
	long := "This is a command with flags for testing purposes"
	runE := func(_ *cobra.Command, _ []string) error { return nil }
	flagCalled := false
	setupFlags := func(cmd *cobra.Command) {
		flagCalled = true

		cmd.Flags().String("test-flag", "", "A test flag")
	}

	// Act
	cmd := factory.NewCobraCommandWithFlags(use, short, long, runE, setupFlags)

	// Assert
	assertBasicCommandProperties(t, cmd, use, short, long)

	if !flagCalled {
		t.Error("expected setupFlags function to be called")
	}

	if cmd.Flags().Lookup("test-flag") == nil {
		t.Error("expected test-flag to be added by setupFlags")
	}
}

// assertBasicCommandProperties is a helper function to avoid code duplication in tests.
func assertBasicCommandProperties(t *testing.T, cmd *cobra.Command, use, short, long string) {
	t.Helper()

	if cmd.Use != use {
		t.Errorf("expected Use to be %q, got %q", use, cmd.Use)
	}

	if cmd.Short != short {
		t.Errorf("expected Short to be %q, got %q", short, cmd.Short)
	}

	if cmd.Long != long {
		t.Errorf("expected Long to be %q, got %q", long, cmd.Long)
	}

	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}

	if cmd.SuggestionsMinimumDistance != factory.SuggestionsMinimumDistance {
		t.Errorf("expected SuggestionsMinimumDistance to be %d, got %d",
			factory.SuggestionsMinimumDistance, cmd.SuggestionsMinimumDistance)
	}
}
