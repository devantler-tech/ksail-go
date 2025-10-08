package kubectl_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/kubectl"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// createTestIOStreams creates IO streams for testing.
func createTestIOStreams() genericiooptions.IOStreams {
	return genericiooptions.IOStreams{
		In:     &bytes.Buffer{},
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
}

func TestNewClient(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)

	require.NotNil(t, client, "expected client to be created")
}

func TestCreateApplyCommand(t *testing.T) {
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

			ioStreams := createTestIOStreams()

			client := kubectl.NewClient(ioStreams)
			cmd := client.CreateApplyCommand(testCase.kubeConfigPath)

			require.NotNil(t, cmd, "expected apply command to be created")
			require.Equal(t, "apply", cmd.Use, "expected command Use to be 'apply'")
			require.Equal(t, "Apply manifests", cmd.Short, "expected command Short description")
			require.Equal(
				t,
				"Apply local Kubernetes manifests to your cluster.",
				cmd.Long,
				"expected command Long description",
			)
		})
	}
}

func TestCreateApplyCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateApplyCommand("/path/to/kubeconfig")

	// Verify that kubectl apply flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("filename"), "expected --filename flag to be present")
	require.NotNil(t, flags.Lookup("recursive"), "expected --recursive flag to be present")
	require.NotNil(t, flags.Lookup("force"), "expected --force flag to be present")
	require.NotNil(t, flags.Lookup("dry-run"), "expected --dry-run flag to be present")
	require.NotNil(t, flags.Lookup("server-side"), "expected --server-side flag to be present")
	require.NotNil(t, flags.Lookup("prune"), "expected --prune flag to be present")
}

func TestCreateGetCommand(t *testing.T) {
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

			ioStreams := createTestIOStreams()

			client := kubectl.NewClient(ioStreams)
			cmd := client.CreateGetCommand(testCase.kubeConfigPath)

			require.NotNil(t, cmd, "expected get command to be created")
			require.Equal(t, "get", cmd.Use, "expected command Use to be 'get'")
			require.Equal(t, "Get resources", cmd.Short, "expected command Short description")
			require.Equal(
				t,
				"Display one or many Kubernetes resources from your cluster.",
				cmd.Long,
				"expected command Long description",
			)
		})
	}
}

func TestCreateGetCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateGetCommand("/path/to/kubeconfig")

	// Verify that kubectl get flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("output"), "expected --output flag to be present")
	require.NotNil(t, flags.Lookup("watch"), "expected --watch flag to be present")
	require.NotNil(
		t,
		flags.Lookup("all-namespaces"),
		"expected --all-namespaces flag to be present",
	)
	require.NotNil(
		t,
		flags.Lookup("field-selector"),
		"expected --field-selector flag to be present",
	)
	require.NotNil(t, flags.Lookup("selector"), "expected --selector flag to be present")
	require.NotNil(t, flags.Lookup("show-labels"), "expected --show-labels flag to be present")
}
