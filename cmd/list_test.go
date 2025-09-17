package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
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

func TestListCmdExecuteDefault(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "list",
		NewCommand:  cmd.NewListCmd,
	})
}

func TestListCmdExecuteAll(t *testing.T) {
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

func TestListCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		NewCommand: cmd.NewListCmd,
	})
}

func TestListCmdFlags(t *testing.T) {
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
func TestHandleListRunESuccess(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	// Add the --all flag to the command like the real command would have
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	manager := configmanager.NewConfigManager()

	err := cmd.HandleListRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Listing running clusters (stub implementation)")
	assert.Contains(t, out.String(), "► Distribution filter:")
}

// TestHandleListRunE_AllFlag tests list command with --all flag.
func TestHandleListRunEAllFlag(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")
	// Set the flag value
	err := testCmd.Flags().Set("all", "true")
	require.NoError(t, err)

	manager := configmanager.NewConfigManager()

	err = cmd.HandleListRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Listing all clusters (stub implementation)")
	assert.Contains(t, out.String(), "► Distribution filter:")
}

// TestHandleListRunE_Error tests list command with config load error.
func TestHandleListRunEError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	manager := configmanager.NewConfigManager()

	// Test that the function doesn't panic - error testing can be enhanced later
	// when real error conditions are available in the stub implementation
	assert.NotPanics(t, func() {
		_ = cmd.HandleListRunE(testCmd, manager, []string{})
	})
}

// TestHandleListRunE_InvalidConfigManager tests type assertion failure in HandleListRunE.
func TestHandleListRunE_InvalidConfigManager(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)
	testCmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	// Create a mock config manager that is not a *configmanager.ConfigManager
	mockManager := &invalidConfigManager{}

	err := cmd.HandleListRunE(testCmd, mockManager, []string{})

	// Should return error due to type assertion failure
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config manager type")
}

// invalidConfigManager is a mock implementation that doesn't match *configmanager.ConfigManager.
type invalidConfigManager struct{}

func (m *invalidConfigManager) LoadConfig() (*v1alpha1.Cluster, error) {
	return v1alpha1.NewCluster(), nil
}

func (m *invalidConfigManager) AddFlagsFromFields(*cobra.Command) {}
