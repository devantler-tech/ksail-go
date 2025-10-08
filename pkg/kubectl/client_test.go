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

func TestCreateCreateCommand(t *testing.T) {
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
			cmd := client.CreateCreateCommand(testCase.kubeConfigPath)

			require.NotNil(t, cmd, "expected create command to be created")
			require.Equal(t, "create", cmd.Use, "expected command Use to be 'create'")
			require.Equal(t, "Create resources", cmd.Short, "expected command Short description")
			require.Equal(
				t,
				"Create Kubernetes resources from files or stdin.",
				cmd.Long,
				"expected command Long description",
			)
		})
	}
}

func TestCreateCreateCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateCreateCommand("/path/to/kubeconfig")

	// Verify that kubectl create flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("filename"), "expected --filename flag to be present")
	require.NotNil(t, flags.Lookup("edit"), "expected --edit flag to be present")
	require.NotNil(t, flags.Lookup("dry-run"), "expected --dry-run flag to be present")
	require.NotNil(t, flags.Lookup("output"), "expected --output flag to be present")
	require.NotNil(t, flags.Lookup("raw"), "expected --raw flag to be present")
}

func TestCreateCreateCommandHasSubcommands(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateCreateCommand("/path/to/kubeconfig")

	// Verify that kubectl create subcommands are present
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands, "expected create command to have subcommands")

	// Check for some common subcommands
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Name()] = true
	}

	expectedSubcommands := []string{
		"deployment",
		"namespace",
		"secret",
		"configmap",
		"service",
		"job",
	}

	for _, expected := range expectedSubcommands {
		require.True(t, subcommandNames[expected], "expected subcommand %q to be present", expected)
	}
}
