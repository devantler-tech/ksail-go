package cmd_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/pkg/config"
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
	t.Parallel()

	version := "1.2.3"
	commit := "abc123"
	date := "2025-08-17"
	cmd := cmd.NewRootCmd(version, commit, date)

	expectedVersion := version + " (Built on " + date + " from Git SHA " + commit + ")"
	if cmd.Version != expectedVersion {
		t.Fatalf("unexpected version string. want %q, got %q", expectedVersion, cmd.Version)
	}
}

func TestExecute_ShowsHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.NewRootCmd("", "", "")
	root.SetOut(&out)

	_ = root.Execute()

	snaps.MatchSnapshot(t, out.String())
}

func TestExecute_ShowsVersion(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.NewRootCmd("1.2.3", "abc123", "2025-08-17")
	root.SetOut(&out)
	root.SetArgs([]string{"--version"})

	_ = root.Execute()

	snaps.MatchSnapshot(t, out.String())
}

// newTestCommand creates a cobra.Command for testing with exhaustive field initialization.
func newTestCommand(use string, runE func(*cobra.Command, []string) error) *cobra.Command {
	return config.NewCobraCommand(use, "", "", func(cmd *cobra.Command, _ *config.Manager, args []string) error {
		return runE(cmd, args)
	})
}

func TestExecute_ReturnsError(t *testing.T) {
	t.Parallel()

	failing := newTestCommand("fail", func(_ *cobra.Command, _ []string) error {
		return errRootTest
	})

	actual := cmd.NewRootCmd("test", "test", "test")
	actual.SetArgs([]string{"fail"})
	actual.AddCommand(failing)

	err := actual.Execute()
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

	actual := cmd.NewRootCmd("test", "test", "test")
	actual.SetArgs([]string{"ok"})
	actual.AddCommand(succeeding)

	err := actual.Execute()
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}
}

func TestExecute_WrapperSuccess(t *testing.T) {
	t.Parallel()

	succeeding := newTestCommand("ok", func(_ *cobra.Command, _ []string) error {
		return nil
	})

	rootCmd := cmd.NewRootCmd("test", "test", "test")
	rootCmd.SetArgs([]string{"ok"})
	rootCmd.AddCommand(succeeding)

	err := cmd.Execute(rootCmd)
	if err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}
}

func TestExecute_WrapperError(t *testing.T) {
	t.Parallel()

	failing := newTestCommand("fail", func(_ *cobra.Command, _ []string) error {
		return errRootTest
	})

	rootCmd := cmd.NewRootCmd("test", "test", "test")
	rootCmd.SetArgs([]string{"fail"})
	rootCmd.AddCommand(failing)

	err := cmd.Execute(rootCmd)
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	if !errors.Is(err, errRootTest) {
		t.Fatalf("Expected error to wrap %v, got %v", errRootTest, err)
	}
}
