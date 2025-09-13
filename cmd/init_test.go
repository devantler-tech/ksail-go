package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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

	if cmd.Short != "Initialize a new project" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestInitCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "init",
		NewCommand:  cmd.NewInitCmd,
	})
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

// TestHandleInitRunE_Success tests successful init command execution.
func TestHandleInitRunE_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := ksail.NewManager()

	err := cmd.HandleInitRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ project initialized successfully")
	assert.Contains(t, out.String(), "► Distribution:")
	assert.Contains(t, out.String(), "► Source directory:")
}

// TestHandleInitRunE_Error tests init command with config load error.
func TestHandleInitRunE_Error(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	mockManager := ksail.NewMockConfigManager[v1alpha1.Cluster](t)
	mockManager.EXPECT().LoadCluster().Return(nil, testutils.ErrTestConfigLoadError)

	err := cmd.HandleInitRunE(testCmd, mockManager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}
