package cmd_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
)

var errRootTest = errors.New("boom")

func TestMain(main *testing.M) { testutils.RunTestMainWithSnapshotCleanup(main) }

func TestNewRootCmdVersionFormatting(t *testing.T) {
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

func TestExecuteShowsHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.NewRootCmd("", "", "")
	root.SetOut(&out)

	_ = root.Execute()

	snaps.MatchSnapshot(t, out.String())
}

func TestExecuteShowsVersion(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.NewRootCmd("1.2.3", "abc123", "2025-08-17")
	root.SetOut(&out)
	root.SetArgs([]string{"--version"})

	_ = root.Execute()

	snaps.MatchSnapshot(t, out.String())
}

func TestClusterCommandShowsHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	root := cmd.NewRootCmd("", "", "")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cluster"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("expected cluster command to show help without error, got %v", err)
	}

	output := out.String()

	generalSnippets := []string{
		"ksail cluster [command]",
		"Available Commands:",
	}

	for _, snippet := range generalSnippets {
		if !strings.Contains(output, snippet) {
			t.Fatalf("expected cluster help output to contain %q, got %q", snippet, output)
		}
	}

	lines := strings.Split(output, "\n")
	expectedCommands := map[string]string{
		"up":     "Start the Kubernetes cluster",
		"down":   "Destroy a cluster",
		"start":  "Start a stopped cluster",
		"stop":   "Stop the Kubernetes cluster",
		"status": "Show status of the Kubernetes cluster",
		"list":   "List clusters",
	}

	for command, description := range expectedCommands {
		found := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, command+" ") && strings.Contains(trimmed, description) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf(
				"expected cluster help output to contain command %q with description %q, got %q",
				command,
				description,
				output,
			)
		}
	}
}

// newTestCommand creates a cobra.Command for testing with exhaustive field initialization.
func newTestCommand(use string, runE func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:  use,
		RunE: runE,
	}
}

func TestExecuteReturnsError(t *testing.T) {
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

func TestExecuteWrapperSuccess(t *testing.T) {
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

func TestExecuteWrapperError(t *testing.T) {
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
