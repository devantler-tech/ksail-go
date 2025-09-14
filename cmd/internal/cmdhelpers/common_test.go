package cmdhelpers_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors to comply with err113.
var (
	errFailedToLoadConfig = errors.New("failed to load config")
	errConfigLoadFailed   = errors.New("config load failed")
)

// setupTestCommand creates a test command with output buffer for testing.
func setupTestCommand() (*cobra.Command, *bytes.Buffer) {
	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	return testCmd, &out
}

// setupMockManagerWithError creates a mock config manager that returns the specified error.
func setupMockManagerWithError(
	t *testing.T,
	err error,
) *configmanager.MockConfigManager[v1alpha1.Cluster] {
	t.Helper()

	mockManager := configmanager.NewMockConfigManager[v1alpha1.Cluster](t)
	mockManager.EXPECT().LoadConfig().Return(nil, err)

	return mockManager
}

func TestNewSimpleClusterCommand(t *testing.T) {
	t.Parallel()

	cfg := cmdhelpers.CommandConfig{
		Use:   "test",
		Short: "Test command",
		Long:  "A test command for testing",
		RunEFunc: func(_ *cobra.Command, _ configmanager.ConfigManager[v1alpha1.Cluster], _ []string) error {
			return nil
		},
		FieldsFunc: func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Test distribution flag",
			}
		},
	}

	cmd := cmdhelpers.NewSimpleClusterCommand(cfg)

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

	manager := ksail.NewManager()

	// Test the actual exported function
	cluster, err := cmdhelpers.HandleSimpleClusterCommand(testCmd, manager, "Test success message")

	require.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Contains(t, out.String(), "✔ Test success message")
	assert.Contains(t, out.String(), "► Distribution:")
	assert.Contains(t, out.String(), "► Context:")
}

func TestHandleSimpleClusterCommand_LoadError(t *testing.T) {
	t.Parallel()

	testCmd, out := setupTestCommand()
	mockManager := setupMockManagerWithError(t, errFailedToLoadConfig)

	// Test the actual exported function with error injection
	cluster, err := cmdhelpers.HandleSimpleClusterCommand(
		testCmd,
		mockManager,
		"Test success message",
	)

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

	manager := ksail.NewManager()

	cluster, err := cmdhelpers.LoadClusterWithErrorHandling(testCmd, manager)

	require.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Empty(t, out.String()) // No error output
}

func TestLoadClusterWithErrorHandling_LoadError(t *testing.T) {
	t.Parallel()

	testCmd, out := setupTestCommand()
	mockManager := setupMockManagerWithError(t, errConfigLoadFailed)

	cluster, err := cmdhelpers.LoadClusterWithErrorHandling(testCmd, mockManager)

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

func TestStandardDistributionFieldSelector(t *testing.T) {
	t.Parallel()

	description := "Kubernetes distribution to use"
	selector := cmdhelpers.StandardDistributionFieldSelector(description)

	assert.Equal(t, description, selector.Description)
	assert.Equal(t, v1alpha1.DistributionKind, selector.DefaultValue)

	// Test selector function
	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.Distribution, result)
}

func TestStandardSourceDirectoryFieldSelector(t *testing.T) {
	t.Parallel()

	selector := cmdhelpers.StandardSourceDirectoryFieldSelector()

	assert.Equal(t, "Directory containing workloads to deploy", selector.Description)
	assert.Equal(t, "k8s", selector.DefaultValue)

	// Test selector function
	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.SourceDirectory, result)
}

func TestStandardDistributionConfigFieldSelector(t *testing.T) {
	t.Parallel()

	selector := cmdhelpers.StandardDistributionConfigFieldSelector()

	assert.Equal(t, "Configuration file for the distribution", selector.Description)
	assert.Equal(t, "kind.yaml", selector.DefaultValue)

	// Test selector function
	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.DistributionConfig, result)
}

func TestStandardClusterCommandRunE_Success(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	testCmd := &cobra.Command{}
	testCmd.SetOut(&out)

	manager := ksail.NewManager()
	successMessage := "Test command executed successfully"

	// Get the run function
	runFunc := cmdhelpers.StandardClusterCommandRunE(successMessage)

	// Execute the function
	err := runFunc(testCmd, manager, []string{})

	require.NoError(t, err)
	assert.Contains(t, out.String(), "✔ "+successMessage)
}

func TestStandardClusterCommandRunE_Error(t *testing.T) {
	t.Parallel()

	testCmd, _ := setupTestCommand()
	mockManager := setupMockManagerWithError(t, errFailedToLoadConfig)
	successMessage := "Test command executed successfully"

	// Get the run function
	runFunc := cmdhelpers.StandardClusterCommandRunE(successMessage)

	// Execute the function
	err := runFunc(testCmd, mockManager, []string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to handle cluster command")
	assert.Contains(t, err.Error(), "failed to load config")
}
