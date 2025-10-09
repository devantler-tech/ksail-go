package k9s_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/k9s"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	t.Run("with custom k9s path", func(t *testing.T) {
		t.Parallel()

		client := k9s.NewClient("/custom/path/to/k9s")
		require.NotNil(t, client, "expected client to be created")
	})

	t.Run("with empty path defaults to k9s", func(t *testing.T) {
		t.Parallel()

		client := k9s.NewClient("")
		require.NotNil(t, client, "expected client to be created")
	})
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

			client := k9s.NewClient("")
			cmd := client.CreateConnectCommand(testCase.kubeConfigPath)

			require.NotNil(t, cmd, "expected command to be created")
			require.Equal(t, "connect", cmd.Use, "expected Use to be 'connect'")
			require.Equal(t, "Connect to cluster with k9s", cmd.Short, "expected Short description")
			require.Contains(t, cmd.Long, "Launch k9s terminal UI", "expected Long description to mention k9s")
			require.True(t, cmd.DisableFlagParsing, "expected DisableFlagParsing to be true")
			require.True(t, cmd.SilenceUsage, "expected SilenceUsage to be true")
		})
	}
}

func TestCreateConnectCommandRunE(t *testing.T) {
	t.Parallel()

	t.Run("command returns error when k9s not found", func(t *testing.T) {
		t.Parallel()

		client := k9s.NewClient("/nonexistent/k9s")
		cmd := client.CreateConnectCommand("")

		// Create buffers for output
		var outBuf, errBuf bytes.Buffer
		cmd.SetOut(&outBuf)
		cmd.SetErr(&errBuf)

		// Execute the command with a cancellable context
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := cmd.ExecuteContext(ctx)
		require.Error(t, err, "expected error when k9s binary not found")

		// Check that it's an error related to exec/file system
		require.Contains(t, err.Error(), "no such file or directory",
			"expected error message to indicate file not found")
	})
}

func TestCreateConnectCommandPassesThroughFlags(t *testing.T) {
	t.Parallel()

	t.Run("command structure supports flag pass-through", func(t *testing.T) {
		t.Parallel()

		client := k9s.NewClient("")
		cmd := client.CreateConnectCommand("/path/to/kubeconfig")

		// Verify that DisableFlagParsing is true, which allows flags to pass through
		require.True(t, cmd.DisableFlagParsing,
			"expected DisableFlagParsing to be true to allow k9s flags to pass through")

		// Verify RunE is set
		require.NotNil(t, cmd.RunE, "expected RunE to be set")
	})
}
