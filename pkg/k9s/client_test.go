package k9s_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/k9s"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	require.NotNil(t, client, "expected client to be created")
}

func TestNewClientWithExecutor(t *testing.T) {
	t.Parallel()

	mockExecutor := k9s.NewMockExecutor(t)
	client := k9s.NewClientWithExecutor(mockExecutor)
	require.NotNil(t, client, "expected client to be created")
}

func TestCreateConnectCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		kubeConfigPath string
	}{
		{
			name:           "with kubeconfig path",
			kubeConfigPath: "/path/to/kubeconfig",
		},
		{
			name:           "without kubeconfig path",
			kubeConfigPath: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := k9s.NewClient()
			cmd := client.CreateConnectCommand(testCase.kubeConfigPath)

			require.NotNil(t, cmd, "expected command to be created")
			require.Equal(t, "connect", cmd.Use, "expected Use to be 'connect'")
			require.Equal(t, "Connect to cluster with k9s", cmd.Short, "expected Short description")
			require.Contains(
				t,
				cmd.Long,
				"Launch k9s terminal UI",
				"expected Long description to mention k9s",
			)
			require.True(t, cmd.SilenceUsage, "expected SilenceUsage to be true")
		})
	}
}

func TestCreateConnectCommandStructure(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	cmd := client.CreateConnectCommand("")

	// Verify RunE is set
	require.NotNil(t, cmd.RunE, "expected RunE to be set")

	// Verify command can be executed (though it will fail without k9s)
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	// We don't actually execute it since k9s would try to start the UI
	// Just verify the command structure is correct
	require.NotNil(t, cmd, "command should be properly structured")
}

func TestCreateConnectCommand_WithKubeconfig(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	kubeConfigPath := "/test/path/to/kubeconfig"
	cmd := client.CreateConnectCommand(kubeConfigPath)

	require.NotNil(t, cmd, "expected command to be created")
	require.NotNil(t, cmd.RunE, "expected RunE to be set")

	// Verify the command metadata
	require.Equal(t, "connect", cmd.Use)
	require.Equal(t, "Connect to cluster with k9s", cmd.Short)
	require.True(t, cmd.SilenceUsage)
}

func TestCreateConnectCommand_WithoutKubeconfig(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	cmd := client.CreateConnectCommand("")

	require.NotNil(t, cmd, "expected command to be created")
	require.NotNil(t, cmd.RunE, "expected RunE to be set")

	// Verify the command metadata
	require.Equal(t, "connect", cmd.Use)
	require.Equal(t, "Connect to cluster with k9s", cmd.Short)
	require.True(t, cmd.SilenceUsage)
}

func TestCreateConnectCommand_WithArgs(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	cmd := client.CreateConnectCommand("/path/to/kubeconfig")

	require.NotNil(t, cmd, "expected command to be created")
	require.NotNil(t, cmd.RunE, "expected RunE to be set")

	// Set some args to pass through
	cmd.SetArgs([]string{"--namespace", "default"})

	// Verify command structure is correct for arg passing
	require.NotNil(t, cmd, "command should accept args")
}

func TestRunK9s_ArgumentHandling(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()

	// Test with kubeconfig path
	cmdWithConfig := client.CreateConnectCommand("/test/kubeconfig")
	require.NotNil(t, cmdWithConfig)
	require.NotNil(t, cmdWithConfig.RunE)

	// Test without kubeconfig path
	cmdWithoutConfig := client.CreateConnectCommand("")
	require.NotNil(t, cmdWithoutConfig)
	require.NotNil(t, cmdWithoutConfig.RunE)

	// Test with args
	cmdWithArgs := client.CreateConnectCommand("/test/kubeconfig")
	cmdWithArgs.SetArgs([]string{"--namespace", "test", "--readonly"})
	require.NotNil(t, cmdWithArgs)
	require.NotNil(t, cmdWithArgs.RunE)
}

func TestRunK9s_WithMockExecutor_WithKubeconfig(t *testing.T) {
	t.Parallel()

	// Save original args and capture them during test
	var capturedArgs []string
	
	mockExecutor := k9s.NewMockExecutor(t)
	mockExecutor.EXPECT().Execute().Run(func() {
		// Capture os.Args when Execute is called
		capturedArgs = make([]string, len(os.Args))
		copy(capturedArgs, os.Args)
	}).Once()
	
	client := k9s.NewClientWithExecutor(mockExecutor)

	cmd := client.CreateConnectCommand("/test/kubeconfig")
	require.NotNil(t, cmd)

	// Execute the command
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Verify os.Args were set correctly
	require.Contains(t, capturedArgs, "k9s")
	require.Contains(t, capturedArgs, "--kubeconfig")
	require.Contains(t, capturedArgs, "/test/kubeconfig")
}

func TestRunK9s_WithMockExecutor_WithoutKubeconfig(t *testing.T) {
	t.Parallel()

	// Save original args and capture them during test
	var capturedArgs []string
	
	mockExecutor := k9s.NewMockExecutor(t)
	mockExecutor.EXPECT().Execute().Run(func() {
		// Capture os.Args when Execute is called
		capturedArgs = make([]string, len(os.Args))
		copy(capturedArgs, os.Args)
	}).Once()
	
	client := k9s.NewClientWithExecutor(mockExecutor)

	cmd := client.CreateConnectCommand("")
	require.NotNil(t, cmd)

	// Execute the command
	err := cmd.RunE(cmd, []string{})
	require.NoError(t, err)

	// Verify os.Args only contains k9s (no kubeconfig)
	require.Contains(t, capturedArgs, "k9s")
	require.NotContains(t, capturedArgs, "--kubeconfig")
}

func TestRunK9s_WithMockExecutor_WithAdditionalArgs(t *testing.T) {
	t.Parallel()

	// Save original args and capture them during test
	var capturedArgs []string
	
	mockExecutor := k9s.NewMockExecutor(t)
	mockExecutor.EXPECT().Execute().Run(func() {
		// Capture os.Args when Execute is called
		capturedArgs = make([]string, len(os.Args))
		copy(capturedArgs, os.Args)
	}).Once()
	
	client := k9s.NewClientWithExecutor(mockExecutor)

	cmd := client.CreateConnectCommand("/test/kubeconfig")
	require.NotNil(t, cmd)

	// Execute the command with additional arguments
	err := cmd.RunE(cmd, []string{"--namespace", "default", "--readonly"})
	require.NoError(t, err)

	// Verify os.Args contains all arguments
	require.Contains(t, capturedArgs, "k9s")
	require.Contains(t, capturedArgs, "--kubeconfig")
	require.Contains(t, capturedArgs, "/test/kubeconfig")
	require.Contains(t, capturedArgs, "--namespace")
	require.Contains(t, capturedArgs, "default")
	require.Contains(t, capturedArgs, "--readonly")
}

func TestRunK9s_OsArgsRestored(t *testing.T) {
	t.Parallel()

	mockExecutor := k9s.NewMockExecutor(t)
	mockExecutor.EXPECT().Execute().Once()
	
	client := k9s.NewClientWithExecutor(mockExecutor)

	// Save original os.Args
	originalArgs := make([]string, len(os.Args))
	copy(originalArgs, os.Args)

	cmd := client.CreateConnectCommand("/test/kubeconfig")
	err := cmd.RunE(cmd, []string{"--namespace", "test"})
	require.NoError(t, err)

	// Verify os.Args is restored after execution
	require.Equal(t, originalArgs, os.Args, "expected os.Args to be restored")
}
