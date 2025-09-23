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

func TestNewReconcileCmd(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandCreation(t, testutils.SimpleCommandTestData{
		CommandName:   "reconcile",
		NewCommand:    cmd.NewReconcileCmd,
		ExpectedUse:   "reconcile",
		ExpectedShort: "Reconcile workloads in the cluster",
	})
}

func TestReconcileCmdExecute(t *testing.T) {
	t.Parallel()

	reconcileCmd := cmd.NewReconcileCmd()

	err := reconcileCmd.Execute()

	// Expect a validation error because no valid configuration is provided
	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration validation failed")
}

func TestReconcileCmdHelp(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
}

// TestHandleReconcileRunE_Success tests successful reconcile command execution.
func TestHandleReconcileRunESuccess(t *testing.T) {
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

	err := cmd.HandleReconcileRunE(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ Workloads reconciled successfully (stub implementation)")
	assert.Contains(t, out.String(), "► Reconciliation tool:")
	assert.Contains(t, out.String(), "► Source directory:")
	assert.Contains(t, out.String(), "► Context:")
}

// TestHandleReconcileRunE_Error tests reconcile command with config load error.
func TestHandleReconcileRunEError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := configmanager.NewConfigManager()

	// Test that the function doesn't panic - error testing can be enhanced later
	assert.NotPanics(t, func() {
		_ = cmd.HandleReconcileRunE(testCmd, manager, []string{})
	})
}
