package k9s_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/k9s"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := k9s.NewClient()
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
