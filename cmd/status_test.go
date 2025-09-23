package cmd_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
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

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := configmanager.NewConfigManager(
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to use",
			DefaultValue: v1alpha1.DistributionKind,
		},
	)

	err := cmd.HandleStatusRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Cluster status: Running (stub implementation)")
	assert.Contains(t, out.String(), "► Context:")
	assert.Contains(t, out.String(), "► Kubeconfig:")
}
