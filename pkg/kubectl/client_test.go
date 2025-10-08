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

func TestCreateEditCommand(t *testing.T) {
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
			cmd := client.CreateEditCommand(testCase.kubeConfigPath)

			require.NotNil(t, cmd, "expected edit command to be created")
			require.Equal(t, "edit", cmd.Use, "expected command Use to be 'edit'")
			require.Equal(t, "Edit a resource", cmd.Short, "expected command Short description")
			require.Equal(
				t,
				"Edit a Kubernetes resource from the default editor.",
				cmd.Long,
				"expected command Long description",
			)
		})
	}
}

func TestCreateEditCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateEditCommand("/path/to/kubeconfig")

	// Verify that kubectl edit flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("filename"), "expected --filename flag to be present")
	require.NotNil(t, flags.Lookup("output"), "expected --output flag to be present")
	require.NotNil(t, flags.Lookup("output-patch"), "expected --output-patch flag to be present")
	require.NotNil(
		t,
		flags.Lookup("windows-line-endings"),
		"expected --windows-line-endings flag to be present",
	)
	require.NotNil(t, flags.Lookup("validate"), "expected --validate flag to be present")
}
