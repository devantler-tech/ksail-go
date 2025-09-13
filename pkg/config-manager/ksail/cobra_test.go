package ksail_test

import (
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestNewCobraCommand_BasicFunctionality tests basic NewCobraCommand functionality.
func TestNewCobraCommand_BasicFunctionality(t *testing.T) {
	t.Parallel()

	var (
		runECalled      bool
		receivedManager *ksail.Manager
		receivedCmd     *cobra.Command
		receivedArgs    []string
	)

	runE := func(cmd *cobra.Command, manager *ksail.Manager, args []string) error {
		runECalled = true
		receivedManager = manager
		receivedCmd = cmd
		receivedArgs = args

		return nil
	}

	cmd := ksail.NewCobraCommand(
		"test",
		"Test command",
		"This is a test command",
		runE,
	)

	require.NotNil(t, cmd)
	assert.Equal(t, "test", cmd.Use)
	assert.Equal(t, "Test command", cmd.Short)
	assert.Equal(t, "This is a test command", cmd.Long)
	assert.Equal(t, ksail.SuggestionsMinimumDistance, cmd.SuggestionsMinimumDistance)

	// Test RunE function
	testArgs := []string{"arg1", "arg2"}
	err := cmd.RunE(cmd, testArgs)
	require.NoError(t, err)

	assert.True(t, runECalled)
	assert.NotNil(t, receivedManager)
	assert.Equal(t, cmd, receivedCmd)
	assert.Equal(t, testArgs, receivedArgs)
}

// TestNewCobraCommand_WithFieldSelectors tests NewCobraCommand with field selectors.
func TestNewCobraCommand_WithFieldSelectors(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
			"Kubernetes distribution",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s",
			"Source directory",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"",
			"Kubernetes context",
		),
	}

	var receivedManager *ksail.Manager

	runE := func(_ *cobra.Command, manager *ksail.Manager, _ []string) error {
		receivedManager = manager

		return nil
	}

	cmd := ksail.NewCobraCommand(
		"init",
		"Initialize cluster",
		"Initialize a new Kubernetes cluster",
		runE,
		fieldSelectors...,
	)

	require.NotNil(t, cmd)
	assert.True(
		t,
		cmd.Flags().HasFlags(),
		"Command should have flags when field selectors are provided",
	)

	// Check that expected flags exist
	expectedFlags := []string{"distribution", "source-directory", "context"}
	for _, flagName := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		assert.NotNil(t, flag, "Flag %s should exist", flagName)
	}

	// Test RunE function to ensure manager is passed correctly
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	require.NotNil(t, receivedManager)
}

// TestNewCobraCommand_NoFieldSelectors tests NewCobraCommand without field selectors.
func TestNewCobraCommand_NoFieldSelectors(t *testing.T) {
	t.Parallel()

	var receivedManager *ksail.Manager

	runE := func(_ *cobra.Command, manager *ksail.Manager, _ []string) error {
		receivedManager = manager

		return nil
	}

	cmd := ksail.NewCobraCommand(
		"status",
		"Show status",
		"Show cluster status",
		runE,
	)

	require.NotNil(t, cmd)
	assert.False(
		t,
		cmd.Flags().HasFlags(),
		"Command should not have flags when no field selectors are provided",
	)

	// Test RunE function
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	require.NotNil(
		t,
		receivedManager,
		"Manager should still be provided even without field selectors",
	)
}

// TestNewCobraCommand_AllFieldTypes tests NewCobraCommand with all supported field types.
//
//nolint:funlen // Testing all field types in cobra command requires comprehensive test cases
func TestNewCobraCommand_AllFieldTypes(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"test-cluster",
			"Cluster name",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
			"Kubernetes distribution",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s",
			"Source directory",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"default-context",
			"Kubernetes context",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			"~/.kube/config",
			"Kubeconfig path",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			metav1.Duration{Duration: 5 * time.Minute},
			"Connection timeout",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool },
			v1alpha1.ReconciliationToolFlux,
			"Reconciliation tool",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.CNI },
			v1alpha1.CNICilium,
			"CNI plugin",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.CSI },
			v1alpha1.CSILocalPathStorage,
			"CSI driver",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.IngressController },
			v1alpha1.IngressControllerTraefik,
			"Ingress controller",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.GatewayController },
			v1alpha1.GatewayControllerTraefik,
			"Gateway controller",
		),
	}

	runE := func(_ *cobra.Command, _ *ksail.Manager, _ []string) error {
		return nil
	}

	cmd := ksail.NewCobraCommand(
		"comprehensive",
		"Comprehensive test",
		"Test command with all field types",
		runE,
		fieldSelectors...,
	)

	require.NotNil(t, cmd)
	assert.True(t, cmd.Flags().HasFlags(), "Command should have flags")

	// Check that all expected flags exist
	expectedFlags := map[string]string{
		"name":                "Cluster name",
		"distribution":        "Kubernetes distribution",
		"source-directory":    "Source directory",
		"context":             "Kubernetes context",
		"kubeconfig":          "Kubeconfig path",
		"timeout":             "Connection timeout",
		"reconciliation-tool": "Reconciliation tool",
		"cni":                 "CNI plugin",
		"csi":                 "CSI driver",
		"ingress-controller":  "Ingress controller",
		"gateway-controller":  "Gateway controller",
	}

	for flagName, expectedUsage := range expectedFlags {
		flag := cmd.Flags().Lookup(flagName)
		require.NotNil(t, flag, "Flag %s should exist", flagName)
		assert.Equal(
			t,
			expectedUsage,
			flag.Usage,
			"Flag %s should have correct usage text",
			flagName,
		)
	}

	// Test that shorthand flags are set correctly
	shorthandFlags := map[string]string{
		"distribution":        "d",
		"context":             "c",
		"kubeconfig":          "k",
		"timeout":             "t",
		"source-directory":    "s",
		"reconciliation-tool": "r",
	}

	for flagName, expectedShorthand := range shorthandFlags {
		flag := cmd.Flags().Lookup(flagName)
		require.NotNil(t, flag, "Flag %s should exist", flagName)
		assert.Equal(
			t,
			expectedShorthand,
			flag.Shorthand,
			"Flag %s should have shorthand %s",
			flagName,
			expectedShorthand,
		)
	}
}

// TestNewCobraCommand_RunEErrorHandling tests error handling in RunE.
func TestNewCobraCommand_RunEErrorHandling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		runE        func(*cobra.Command, *ksail.Manager, []string) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "Successful execution",
			runE: func(_ *cobra.Command, _ *ksail.Manager, _ []string) error {
				return nil
			},
			expectError: false,
		},
		{
			name: "Error in execution",
			runE: func(_ *cobra.Command, _ *ksail.Manager, _ []string) error {
				return assert.AnError
			},
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := ksail.NewCobraCommand(
				"error-test",
				"Error test",
				"Test error handling",
				testCase.runE,
			)

			err := cmd.RunE(cmd, []string{})

			if testCase.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewCobraCommand_ManagerConfiguration tests that the manager is configured correctly.
func TestNewCobraCommand_ManagerConfiguration(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"test-cluster",
			"Cluster name",
		),
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionK3d,
			"Kubernetes distribution",
		),
	}

	var receivedManager *ksail.Manager

	runE := func(_ *cobra.Command, manager *ksail.Manager, _ []string) error {
		receivedManager = manager

		return nil
	}

	cmd := ksail.NewCobraCommand(
		"config-test",
		"Configuration test",
		"Test manager configuration",
		runE,
		fieldSelectors...,
	)

	// Execute the command
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	require.NotNil(t, receivedManager)

	// Test that the manager has the correct configuration
	assert.NotNil(t, receivedManager.Config)
	assert.NotNil(t, receivedManager.GetViper())

	// Test that we can load configuration
	config, err := receivedManager.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Config should have default values applied
	assert.Equal(t, "test-cluster", config.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionK3d, config.Spec.Distribution)
}

// TestNewCobraCommand_FlagBinding tests that flags are properly bound to viper.
func TestNewCobraCommand_FlagBinding(t *testing.T) {
	t.Parallel()

	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		ksail.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
			"Kubernetes distribution",
		),
	}

	var receivedManager *ksail.Manager

	runE := func(_ *cobra.Command, manager *ksail.Manager, _ []string) error {
		receivedManager = manager

		return nil
	}

	cmd := ksail.NewCobraCommand(
		"binding-test",
		"Binding test",
		"Test flag binding",
		runE,
		fieldSelectors...,
	)

	// Set a flag value
	_ = cmd.Flags().Set("distribution", "K3d")

	// Execute the command
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	require.NotNil(t, receivedManager)

	// Test that the flag value is reflected in viper
	viperInstance := receivedManager.GetViper()
	assert.Equal(t, "K3d", viperInstance.GetString("distribution"))
}

// TestNewCobraCommand_EmptyFieldSelectors tests NewCobraCommand with empty field selectors.
func TestNewCobraCommand_EmptyFieldSelectors(t *testing.T) {
	t.Parallel()

	var receivedManager *ksail.Manager

	runE := func(_ *cobra.Command, manager *ksail.Manager, _ []string) error {
		receivedManager = manager

		return nil
	}

	// Pass empty slice of field selectors
	cmd := ksail.NewCobraCommand(
		"empty-test",
		"Empty test",
		"Test with empty field selectors",
		runE,
		[]ksail.FieldSelector[v1alpha1.Cluster]{}...,
	)

	require.NotNil(t, cmd)
	assert.False(
		t,
		cmd.Flags().HasFlags(),
		"Command should not have flags when field selectors are empty",
	)

	// Execute the command
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)
	require.NotNil(t, receivedManager)
}
