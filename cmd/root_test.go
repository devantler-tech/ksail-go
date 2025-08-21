package cmd_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
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
	snaps.MatchSnapshot(test, out.String())
}

func TestExecute_ReturnsError(t *testing.T) {
	// Arrange
	t.Parallel()

	failing := &cobra.Command{
		Use: "fail",
		RunE: func(_ *cobra.Command, _ []string) error {
			return errRootTest
		},
	}

	// Act
	err := cmd.Execute(failing)

	// Assert
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected error 'boom', got %v", err)
	}
}

func TestExecute_ReturnsNil(t *testing.T) {
	// Arrange
	t.Parallel()

	succeeding := &cobra.Command{
		Use: "ok",
		RunE: func(_ *cobra.Command, _ []string) error {
			return nil
		},
	}

	// Act
	err := cmd.Execute(succeeding)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
