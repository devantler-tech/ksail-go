package helpers_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogClusterInfo(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	cmd := &cobra.Command{}
	cmd.SetOut(&out)

	fields := []struct {
		Label string
		Value string
	}{
		{"Distribution", "Kind"},
		{"Context", "kind-kind"},
	}

	for _, field := range fields {
		cmd.Printf("► %s: %s\n", field.Label, field.Value)
	}

	assert.Contains(t, out.String(), "► Distribution: Kind")
	assert.Contains(t, out.String(), "► Context: kind-kind")
}

func TestStandardDistributionFieldSelector(t *testing.T) {
	t.Parallel()

	selector := configmanager.StandardDistributionFieldSelector()

	assert.Equal(t, "Kubernetes distribution to use", selector.Description)
	assert.Equal(t, v1alpha1.DistributionKind, selector.DefaultValue)

	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.Distribution, result)
}

func TestStandardSourceDirectoryFieldSelector(t *testing.T) {
	t.Parallel()

	selector := configmanager.StandardSourceDirectoryFieldSelector()

	assert.Equal(t, "Directory containing workloads to deploy", selector.Description)
	assert.Equal(t, "k8s", selector.DefaultValue)

	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.SourceDirectory, result)
}

func TestStandardDistributionConfigFieldSelector(t *testing.T) {
	t.Parallel()

	selector := configmanager.StandardDistributionConfigFieldSelector()

	assert.Equal(t, "Configuration file for the distribution", selector.Description)
	assert.Equal(t, "kind.yaml", selector.DefaultValue)

	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.DistributionConfig, result)
}

func TestStandardContextFieldSelector(t *testing.T) {
	t.Parallel()

	selector := configmanager.StandardContextFieldSelector()

	assert.Equal(t, "Kubernetes context of cluster", selector.Description)
	assert.Equal(t, "kind-kind", selector.DefaultValue)

	cluster := &v1alpha1.Cluster{}
	result := selector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.Connection.Context, result)
}

func TestNewCobraCommand(t *testing.T) {
	t.Parallel()

	var (
		runECalled      bool
		receivedManager *configmanager.ConfigManager
		receivedCmd     *cobra.Command
		receivedArgs    []string
	)

	runE := func(cmd *cobra.Command, manager *configmanager.ConfigManager, args []string) error {
		runECalled = true
		receivedManager = manager
		receivedCmd = cmd
		receivedArgs = args

		return nil
	}

	cmd := helpers.NewCobraCommand(
		"test",
		"Test command",
		"This is a test command",
		runE,
	)

	require.NotNil(t, cmd)
	assert.Equal(t, "test", cmd.Use)
	assert.Equal(t, "Test command", cmd.Short)
	assert.Equal(t, "This is a test command", cmd.Long)
	assert.Equal(t, helpers.SuggestionsMinimumDistance, cmd.SuggestionsMinimumDistance)

	testArgs := []string{"arg1", "arg2"}
	err := cmd.RunE(cmd, testArgs)

	require.NoError(t, err)
	assert.True(t, runECalled)
	assert.NotNil(t, receivedManager)
	assert.Equal(t, cmd, receivedCmd)
	assert.Equal(t, testArgs, receivedArgs)
}

func TestNewCobraCommandWithMultipleFieldSelectors(t *testing.T) {
	t.Parallel()

	var (
		runECalled   bool
		receivedArgs []string
	)

	runE := func(_ *cobra.Command, _ *configmanager.ConfigManager, args []string) error {
		runECalled = true
		receivedArgs = args

		return nil
	}

	cmd := helpers.NewCobraCommand(
		"multi-test",
		"Multi field test command",
		"This command tests multiple field selectors",
		runE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardSourceDirectoryFieldSelector(),
		configmanager.StandardDistributionConfigFieldSelector(),
	)

	require.NotNil(t, cmd)
	assert.Equal(t, "multi-test", cmd.Use)
	assert.Equal(t, "Multi field test command", cmd.Short)
	assert.Equal(t, "This command tests multiple field selectors", cmd.Long)
	assert.NotNil(t, cmd.Flags())

	testArgs := []string{"arg1", "arg2", "arg3"}
	err := cmd.RunE(cmd, testArgs)

	require.NoError(t, err)
	assert.True(t, runECalled)
	assert.Equal(t, testArgs, receivedArgs)
}

func TestNewCobraCommandWithNoFieldSelectors(t *testing.T) {
	t.Parallel()

	var runECalled bool

	runE := func(_ *cobra.Command, _ *configmanager.ConfigManager, _ []string) error {
		runECalled = true
		return nil
	}

	cmd := helpers.NewCobraCommand(
		"no-fields",
		"No fields command",
		"This command has no field selectors",
		runE,
	)

	require.NotNil(t, cmd)
	assert.Equal(t, "no-fields", cmd.Use)

	err := cmd.RunE(cmd, []string{})

	require.NoError(t, err)
	assert.True(t, runECalled)
}

func TestStandardFieldSelectorsComprehensive(t *testing.T) {
	t.Parallel()

	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			Distribution:       v1alpha1.DistributionK3d,
			DistributionConfig: "k3d.yaml",
			SourceDirectory:    "manifests",
			Connection: v1alpha1.Connection{
				Context: "k3d-test-cluster",
			},
		},
	}

	distSelector := configmanager.StandardDistributionFieldSelector()
	distResult := distSelector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.Distribution, distResult)
	assert.Equal(t, "Kubernetes distribution to use", distSelector.Description)
	assert.Equal(t, v1alpha1.DistributionKind, distSelector.DefaultValue)

	srcSelector := configmanager.StandardSourceDirectoryFieldSelector()
	srcResult := srcSelector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.SourceDirectory, srcResult)
	assert.Equal(t, "Directory containing workloads to deploy", srcSelector.Description)
	assert.Equal(t, "k8s", srcSelector.DefaultValue)

	configSelector := configmanager.StandardDistributionConfigFieldSelector()
	configResult := configSelector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.DistributionConfig, configResult)
	assert.Equal(t, "Configuration file for the distribution", configSelector.Description)
	assert.Equal(t, "kind.yaml", configSelector.DefaultValue)

	contextSelector := configmanager.StandardContextFieldSelector()
	contextResult := contextSelector.Selector(cluster)
	assert.Equal(t, &cluster.Spec.Connection.Context, contextResult)
	assert.Equal(t, "Kubernetes context of cluster", contextSelector.Description)
	assert.Equal(t, "kind-kind", contextSelector.DefaultValue)
}
