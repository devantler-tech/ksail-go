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

func TestMain(m *testing.M) {
	v := m.Run()
	// Sort snapshots to ensure deterministic order and clean obsolete ones
	snaps.Clean(m, snaps.CleanOpts{Sort: true})
	os.Exit(v)
}

func TestNewRootCmd_VersionFormatting(test *testing.T) {
	// Arrange
	test.Parallel()

	// Act
	version := "1.2.3"
	commit := "abc123"
	date := "2025-08-17"
	cmd := cmd.NewRootCmd(version, commit, date)

	// Assert
	expectedVersion := version + " (Built on " + date + " from Git SHA " + commit + ")"
	if cmd.Version != expectedVersion {
		test.Fatalf("unexpected version string. want %q, got %q", expectedVersion, cmd.Version)
	}
}

func TestRootCmd_NoArgs_ShowsHelp(test *testing.T) {
	// Arrange
	var out bytes.Buffer

	test.Parallel()

	root := cmd.NewRootCmd("", "", "")
	root.SetOut(&out)

	// Act
	_ = root.Execute()

	// Assert
	snaps.MatchSnapshot(test, out.String())
}

func TestExecute_PropagatesError(test *testing.T) {
	// Arrange
	test.Parallel()

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
		test.Fatalf("expected error 'boom', got %v", err)
	}
}
