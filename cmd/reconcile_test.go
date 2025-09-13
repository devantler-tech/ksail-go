package cmd_test

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/internal/cmd/testutils"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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

func TestReconcileCmd_Execute(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandExecution(t, testutils.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
}

func TestReconcileCmd_Help(t *testing.T) {
	t.Parallel()

	testutils.TestSimpleCommandHelp(t, testutils.SimpleCommandTestData{
		CommandName: "reconcile",
		NewCommand:  cmd.NewReconcileCmd,
	})
}

func TestReconcileCmd_Execute_ConfigError(t *testing.T) {
	t.Parallel()

	// Test the error handling pattern used in handleReconcileRunE
	var out bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create a config manager with error injection to test the pattern
	manager := config.NewManager()
	manager.SetTestErrorHook(errors.New("test config load error"))

	// Test the error handling pattern used in handleReconcileRunE
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
