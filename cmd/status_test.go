package cmd_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatusCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "status",
		NewCommand:    cmd.NewStatusCmd,
		ExpectedUse:   "status",
		ExpectedShort: "Show status of the Kubernetes cluster",
	})
}

func TestStatusCmdExecute(t *testing.T) {
	t.Parallel()

	statusCmd := cmd.NewStatusCmd()

	err := statusCmd.Execute()

	// Expect a validation error because no valid configuration is provided
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestStatusCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "status",
		NewCommand:  cmd.NewStatusCmd,
	})
}

// TestHandleStatusRunE_Success tests successful status command execution.
func TestHandleStatusRunESuccess(t *testing.T) {
	t.Parallel()

	testCmd, out := testutils.SetupCommandWithOutput()
	manager := testutils.CreateDefaultConfigManager()

	err := cmd.HandleStatusRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Cluster status: Running (stub implementation)")
	assert.Contains(t, out.String(), "► Context:")
	assert.Contains(t, out.String(), "► Kubeconfig:")
}
