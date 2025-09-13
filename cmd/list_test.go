package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewListCmd(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewListCmd()

	if cmd == nil {
		t.Fatal("expected command to be created")
	}

	if cmd.Use != "list" {
		t.Fatalf("expected Use to be 'list', got %q", cmd.Use)
	}

	if cmd.Short != "List clusters" {
		t.Fatalf("expected Short description, got %q", cmd.Short)
	}
}

func TestListCmd_Execute_Default(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "list",
		NewCommand:  cmd.NewListCmd,
	})
}

func TestListCmd_Execute_All(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "list",
		NewCommand: func() *cobra.Command {
			cmd := cmd.NewListCmd()
			cmd.SetArgs([]string{"--all"})

			return cmd
		},
	})
}

func TestListCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		NewCommand: cmd.NewListCmd,
	})
}

func TestListCmd_Flags(t *testing.T) {
	t.Parallel()

	cmd := cmd.NewListCmd()

	// Act & Assert
	allFlag := cmd.Flags().Lookup("all")
	if allFlag == nil {
		t.Fatal("expected all flag to exist")
	}

	if allFlag.DefValue != "false" {
		t.Fatalf("expected all flag default to be 'false', got %q", allFlag.DefValue)
	}
}

// TestHandleListRunE_Success tests successful list command execution.
func TestHandleListRunE_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	// Add the --all flag to the command like the real command would have
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	manager := config.NewManager()

	err := cmd.HandleListRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Listing running clusters (stub implementation)")
	assert.Contains(t, out.String(), "► Distribution filter:")
}

// TestHandleListRunE_AllFlag tests list command with --all flag.
func TestHandleListRunE_AllFlag(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")
	// Set the flag value
	err := testCmd.Flags().Set("all", "true")
	require.NoError(t, err)

	manager := config.NewManager()

	err = cmd.HandleListRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Listing all clusters (stub implementation)")
	assert.Contains(t, out.String(), "► Distribution filter:")
}

// TestHandleListRunE_Error tests list command with config load error.
func TestHandleListRunE_Error(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	mockManager := config.NewMockConfigManager(t)
	// Create a real viper instance for the BindPFlag call
	viperInstance := config.NewManager().GetViper()
	mockManager.EXPECT().GetViper().Return(viperInstance).Once()
	mockManager.EXPECT().LoadCluster().Return(nil, testutils.ErrTestConfigLoadError).Once()

	err := cmd.HandleListRunE(testCmd, mockManager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "test config load error")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}
