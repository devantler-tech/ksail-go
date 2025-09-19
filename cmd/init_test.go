package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestNewInitCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewInitCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "init" {
		t.Fatalf("expected Use to be 'init', got %q", cmd.Use)
	}

	if cmd.Short != "Scaffold a new project" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestInitCmdExecute(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Create command and set temporary output directory
	cmd := cmd.NewInitCmd()
	cmd.SetArgs([]string{"--output", tempDir})

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	require.NoError(t, err)

	// Use snapshot testing for the output
	snaps.MatchSnapshot(t, out.String())
}

func TestInitCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		NewCommand: cmd.NewInitCmd,
	})
}

func TestInitCmdFlags(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewInitCmd()

	// Act & Assert
	distributionFlag := cmd.Flags().Lookup("distribution")
	if distributionFlag == nil {
		t.Fatal("expected distribution flag to exist")
	}

	// Verify that CLI flags show appropriate defaults for better UX
	// Distribution should show its default value
	if distributionFlag.DefValue != "Kind" {
		t.Fatalf(
			"expected distribution default to be 'Kind' for help display, got %q",
			distributionFlag.DefValue,
		)
	}

	sourceDirectoryFlag := cmd.Flags().Lookup("source-directory")
	if sourceDirectoryFlag == nil {
		t.Fatal("expected source-directory flag to exist")
	}

	// Source directory should show its default value
	if sourceDirectoryFlag.DefValue != "k8s" {
		t.Fatalf(
			"expected source-directory default to be 'k8s' for help display, got %q",
			sourceDirectoryFlag.DefValue,
		)
	}
}

// TestHandleInitRunE_Success tests successful init command execution.
func TestHandleInitRunESuccess(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create the command and set output directory
	testCmd := cmd.NewInitCmd()
	testCmd.SetArgs([]string{"--output", tempDir})

	// Execute the command
	err := testCmd.Execute()

	// Verify execution was successful
	require.NoError(t, err)

	// Verify files were created
	expectedFiles := []string{
		"ksail.yaml",
		"kind.yaml",
		".sops.yaml",
		"k8s/kustomization.yaml",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(tempDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("expected file %s to be created", file)
		}
	}
}

// Error testing removed - will be reimplemented with concrete types
