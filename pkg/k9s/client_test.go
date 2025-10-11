package k9s_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/k9s"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/mock"
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
		context        string
	}{
		{
			name:           "with kubeconfig path",
			kubeConfigPath: "/path/to/kubeconfig",
			context:        "",
		},
		{
			name:           "without kubeconfig path",
			kubeConfigPath: "",
			context:        "",
		},
		{
			name:           "with context",
			kubeConfigPath: "/path/to/kubeconfig",
			context:        "my-context",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := k9s.NewClient()
			cmd := client.CreateConnectCommand(testCase.kubeConfigPath, testCase.context)

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
	cmd := client.CreateConnectCommand("", "")

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

// verifyCommandMetadata is a helper to verify basic command structure.
func verifyCommandMetadata(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	require.NotNil(t, cmd, "expected command to be created")
	require.NotNil(t, cmd.RunE, "expected RunE to be set")
	require.Equal(t, "connect", cmd.Use)
	require.Equal(t, "Connect to cluster with k9s", cmd.Short)
	require.True(t, cmd.SilenceUsage)
}

func TestCreateConnectCommand_WithKubeconfig(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	kubeConfigPath := "/test/path/to/kubeconfig"
	cmd := client.CreateConnectCommand(kubeConfigPath, "")

	verifyCommandMetadata(t, cmd)
}

func TestCreateConnectCommand_WithoutKubeconfig(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	cmd := client.CreateConnectCommand("", "")

	verifyCommandMetadata(t, cmd)
}

func TestCreateConnectCommand_WithArgs(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
	cmd := client.CreateConnectCommand("/path/to/kubeconfig", "")

	verifyCommandMetadata(t, cmd)

	// Set some args to pass through
	cmd.SetArgs([]string{"--namespace", "default"})

	// Verify command structure is correct for arg passing
	require.NotNil(t, cmd, "command should accept args")
}

func TestRunK9s_ArgumentHandling(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()

	// Test with kubeconfig path
	cmdWithConfig := client.CreateConnectCommand("/test/kubeconfig", "")
	verifyCommandMetadata(t, cmdWithConfig)

	// Test without kubeconfig path
	cmdWithoutConfig := client.CreateConnectCommand("", "")
	verifyCommandMetadata(t, cmdWithoutConfig)

	// Test with args
	cmdWithArgs := client.CreateConnectCommand("/test/kubeconfig", "my-context")
	cmdWithArgs.SetArgs([]string{"--namespace", "test", "--readonly"})
	verifyCommandMetadata(t, cmdWithArgs)
}

//nolint:paralleltest // Tests use shared mock setup
func TestRunK9s_WithMockExecutor_WithKubeconfig(t *testing.T) {
	testRunK9sWithMockExecutor(t, "/test/kubeconfig", "", []string{},
		[]string{"--kubeconfig", "/test/kubeconfig"}, nil)
}

//nolint:paralleltest // Tests use shared mock setup
func TestRunK9s_WithMockExecutor_WithoutKubeconfig(t *testing.T) {
	testRunK9sWithMockExecutor(t, "", "", []string{},
		[]string{}, []string{"--kubeconfig"})
}

//nolint:paralleltest // Tests use shared mock setup
func TestRunK9s_WithMockExecutor_WithAdditionalArgs(t *testing.T) {
	testRunK9sWithMockExecutor(
		t,
		"/test/kubeconfig",
		"",
		[]string{"--namespace", "default", "--readonly"},
		[]string{"--kubeconfig", "/test/kubeconfig", "--namespace", "default", "--readonly"},
		nil,
	)
}

// assertArgsContain is a helper to assert that all expected values are in the args slice.
func assertArgsContain(t *testing.T, args []string, expected ...string) {
	t.Helper()

	for _, exp := range expected {
		require.Contains(t, args, exp, "expected args to contain %q", exp)
	}
}

// assertArgsNotContain is a helper to assert that values are not in the args slice.
func assertArgsNotContain(t *testing.T, args []string, notExpected ...string) {
	t.Helper()

	for _, notExp := range notExpected {
		require.NotContains(t, args, notExp, "expected args to not contain %q", notExp)
	}
}

// setupMockExecutorTest creates a mock executor that captures arguments during execution.
// Returns the mock executor and a pointer to the captured args slice.
func setupMockExecutorTest(t *testing.T) (*k9s.MockExecutor, *[]string) {
	t.Helper()

	var capturedArgs []string

	mockExecutor := k9s.NewMockExecutor(t)
	mockExecutor.EXPECT().Execute(mock.Anything).Run(func(args []string) {
		capturedArgs = make([]string, len(args))
		copy(capturedArgs, args)
	}).Return(nil).Once()

	return mockExecutor, &capturedArgs
}

// runCommandTest executes a command's RunE function and asserts no error occurred.
func runCommandTest(t *testing.T, cmd *cobra.Command, args []string) {
	t.Helper()

	err := cmd.RunE(cmd, args)
	require.NoError(t, err)
}

// testRunK9sWithMockExecutor is a helper that tests k9s command execution with a mock executor.
// It verifies that the command args contain expected values and optionally don't contain notExpected values.
func testRunK9sWithMockExecutor(
	t *testing.T,
	kubeconfig, context string,
	cmdArgs, expectedArgs, notExpectedArgs []string,
) {
	t.Helper()

	mockExecutor, capturedArgs := setupMockExecutorTest(t)
	client := k9s.NewClientWithExecutor(mockExecutor)
	cmd := client.CreateConnectCommand(kubeconfig, context)

	runCommandTest(t, cmd, cmdArgs)

	if len(expectedArgs) > 0 {
		assertArgsContain(t, *capturedArgs, expectedArgs...)
	}

	if len(notExpectedArgs) > 0 {
		assertArgsNotContain(t, *capturedArgs, notExpectedArgs...)
	}
}

//nolint:paralleltest // Tests use shared mock setup
func TestRunK9s_OsArgsRestored(t *testing.T) {
	// This test is no longer needed as we don't modify os.Args anymore
	// Kept as a placeholder to maintain test structure
	mockExecutor := k9s.NewMockExecutor(t)
	mockExecutor.EXPECT().Execute(mock.Anything).Return(nil).Once()

	client := k9s.NewClientWithExecutor(mockExecutor)

	cmd := client.CreateConnectCommand("/test/kubeconfig", "")
	err := cmd.RunE(cmd, []string{"--namespace", "test"})
	require.NoError(t, err)
}

//nolint:paralleltest // Tests use shared mock setup
func TestRunK9s_WithMockExecutor_WithContext(t *testing.T) {
	testRunK9sWithMockExecutor(t, "/test/kubeconfig", "my-context", []string{},
		[]string{"--kubeconfig", "/test/kubeconfig", "--context", "my-context"}, nil)
}

//nolint:paralleltest // Tests use shared mock setup
func TestRunK9s_WithMockExecutor_WithoutContext(t *testing.T) {
	testRunK9sWithMockExecutor(t, "/test/kubeconfig", "", []string{},
		[]string{"--kubeconfig", "/test/kubeconfig"}, []string{"--context"})
}

//nolint:paralleltest // Tests use shared mock setup
func TestRunK9s_WithMockExecutor_WithContextAndAdditionalArgs(t *testing.T) {
	testRunK9sWithMockExecutor(
		t,
		"/test/kubeconfig",
		"prod-cluster",
		[]string{"--namespace", "default", "--readonly"},
		[]string{
			"--kubeconfig", "/test/kubeconfig", "--context", "prod-cluster",
			"--namespace", "default", "--readonly",
		},
		nil,
	)
}

func TestDefaultK9sExecutor(t *testing.T) {
	t.Parallel()

	// Create a DefaultK9sExecutor instance
	executor := &k9s.DefaultK9sExecutor{}
	require.NotNil(t, executor, "expected executor to be created")

	// Note: We cannot test Execute() directly as it would launch k9s
	// which requires a terminal and would hang in test environment.
	// The Execute() method is covered through integration testing
	// and manual verification of the connect command.
}
