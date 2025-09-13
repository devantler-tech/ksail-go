package cmd_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors to comply with err113.
var (
	errFailedToLoadConfig = errors.New("failed to load config")
	errConfigLoadFailed   = errors.New("config load failed")
)

func TestNewSimpleClusterCommand(t *testing.T) {
	t.Parallel()

	cfg := cmd.CommandConfig{
		Use:   "test",
		Short: "Test command",
		Long:  "A test command for testing",
		RunEFunc: func(_ *cobra.Command, _ *config.Manager, _ []string) error {
			return nil
		},
		FieldsFunc: func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Test distribution flag",
			}
		},
	}

	cmd := cmd.NewSimpleClusterCommand(cfg)

	assert.NotNil(t, cmd)
	assert.Equal(t, "test", cmd.Use)
	assert.Equal(t, "Test command", cmd.Short)
	assert.Equal(t, "A test command for testing", cmd.Long)

	// Check that the command has the expected flags
	distributionFlag := cmd.Flags().Lookup("distribution")
	assert.NotNil(t, distributionFlag)
}

func TestHandleSimpleClusterCommand_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()

	// Test the actual exported function
	cluster, err := cmd.HandleSimpleClusterCommand(testCmd, manager, "Test success message")

	require.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Contains(t, out.String(), "✔ Test success message")
	assert.Contains(t, out.String(), "► Distribution:")
	assert.Contains(t, out.String(), "► Context:")
}

func TestHandleSimpleClusterCommand_LoadError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create a config manager with error injection
	manager := config.NewManager()
	manager.SetTestErrorHook(errFailedToLoadConfig)

	// Test the actual exported function with error injection
	cluster, err := cmd.HandleSimpleClusterCommand(testCmd, manager, "Test success message")

	require.Error(t, err)
	assert.Nil(t, cluster)
	assert.Contains(t, err.Error(), "failed to load config")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}

func TestLoadClusterWithErrorHandling_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := config.NewManager()

	cluster, err := cmd.LoadClusterWithErrorHandling(testCmd, manager)

	require.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Empty(t, out.String()) // No error output
}

func TestLoadClusterWithErrorHandling_LoadError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	// Create a config manager with error injection
	manager := config.NewManager()
	manager.SetTestErrorHook(errConfigLoadFailed)

	cluster, err := cmd.LoadClusterWithErrorHandling(testCmd, manager)

	require.Error(t, err)
	assert.Nil(t, cluster)
	assert.Contains(t, err.Error(), "config load failed")
	assert.Contains(t, out.String(), "✗ Failed to load cluster configuration:")
}

func TestLogClusterInfo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := &cobra.Command{}
	cmd.SetOut(&out)

	// Test the logClusterInfo function pattern
	fields := []struct {
		Label string
		Value string
	}{
		{"Distribution", "Kind"},
		{"Context", "kind-ksail-default"},
	}

	for _, field := range fields {
		cmd.Printf("► %s: %s\n", field.Label, field.Value)
	}

	assert.Contains(t, out.String(), "► Distribution: Kind")
	assert.Contains(t, out.String(), "► Context: kind-ksail-default")
}
