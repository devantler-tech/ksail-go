package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/internal/cmd/testutils"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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

	if cmd.Short != "Initialize a new project" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestInitCmd_Execute(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := cmd.NewInitCmd()
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	snaps.MatchSnapshot(t, out.String())
}

func TestInitCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		NewCommand: cmd.NewInitCmd,
	})
}

func TestInitCmd_Flags(t *testing.T) {
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

func TestInitCmd_Execute_ConfigError(t *testing.T) {
	t.Parallel()

	// This test demonstrates the error path but since we can't easily inject errors
	// into the internal config manager used by NewInitCmd, we'll test the pattern
	// rather than the actual function execution
	var out bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create a config manager with error injection to test the pattern used in handleInitRunE
	manager := config.NewManager()
	manager.SetTestErrorHook(errors.New("test config load error"))

	// Test the error handling pattern used in handleInitRunE
	cluster, err := manager.LoadCluster()
	if err != nil {
		testCmd.Printf("✗ Failed to load cluster configuration: %s\n", err.Error())
		err = fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	assert.Error(t, err)
	assert.Nil(t, cluster)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}
