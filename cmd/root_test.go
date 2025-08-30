package cmd_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

var errRootTest = errors.New("boom")

func TestMain(main *testing.M) {
	exitCode := main.Run()

	cleaned, err := snaps.Clean(main, snaps.CleanOpts{Sort: true})
	if err != nil {
		_, _ = os.Stderr.WriteString("failed to clean snapshots: " + err.Error() + "\n")
		os.Exit(1)
	}

	_ = cleaned

	os.Exit(exitCode)
}

func TestNewRootCmd_VersionFormatting(t *testing.T) {
	// Arrange
	t.Parallel()

	// Act
	version := "1.2.3"
	commit := "abc123"
	date := "2025-08-17"
	cmd := cmd.NewRootCmd(version, commit, date)

	// Assert
	expectedVersion := version + " (Built on " + date + " from Git SHA " + commit + ")"
	if cmd.Version != expectedVersion {
		t.Fatalf("unexpected version string. want %q, got %q", expectedVersion, cmd.Version)
	}
}

func TestExecute_ShowsHelp(t *testing.T) {
	// Arrange
	t.Parallel()

	var out bytes.Buffer

	root := cmd.NewRootCmd("", "", "")
	root.SetOut(&out)

	// Act
	_ = root.Execute()

	// Assert
	snaps.MatchSnapshot(t, out.String())
}

// newTestCommand creates a cobra.Command for testing with exhaustive field initialization.
func newTestCommand(use string, runE func(*cobra.Command, []string) error) *cobra.Command {
	return factory.NewCobraCommand(use, "", "", runE)
}

func TestExecute_ReturnsError(t *testing.T) {
	// Arrange
	t.Parallel()

	failing := newTestCommand("fail", func(_ *cobra.Command, _ []string) error {
		return errRootTest
	})

	actual := cmd.NewRootCmd("test", "test", "test")
	actual.SetArgs([]string{"fail"})
	actual.AddCommand(failing)

	// Act
	err := actual.Execute()

	// Assert
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	if !errors.Is(err, errRootTest) {
		t.Fatalf("Expected error to be %v, got %v", errRootTest, err)
	}
}

func TestExecuteWithNonexistentCommand(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewRootCmd("test", "test", "test")
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error but got none")
	}
}

func TestExecuteSuccess(t *testing.T) {
	t.Parallel()

	succeeding := newTestCommand("ok", func(_ *cobra.Command, _ []string) error {
		return nil
	})

	// Act
	actual := cmd.NewRootCmd("test", "test", "test")
	actual.SetArgs([]string{"ok"})
	actual.AddCommand(succeeding)

	err := actual.Execute()

	// Assert
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}
}
