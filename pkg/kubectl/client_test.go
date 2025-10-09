package kubectl_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/kubectl"
	"github.com/spf13/cobra"
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

// testCommandCreation is a helper function to test command creation with various kubeconfig paths.
func testCommandCreation(
	t *testing.T,
	createCmd func(*kubectl.Client, string) *cobra.Command,
	expectedUse string,
	expectedShort string,
	expectedLong string,
) {
	t.Helper()

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
			cmd := createCmd(client, testCase.kubeConfigPath)

			require.NotNil(t, cmd, "expected command to be created")
			require.Equal(t, expectedUse, cmd.Use, "expected command Use to be '%s'", expectedUse)
			require.Equal(t, expectedShort, cmd.Short, "expected command Short description")
			require.Equal(t, expectedLong, cmd.Long, "expected command Long description")
		})
	}
}

func TestCreateApplyCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateApplyCommand(path) },
		"apply",
		"Apply manifests",
		"Apply local Kubernetes manifests to your cluster.",
	)
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

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateEditCommand(path) },
		"edit",
		"Edit a resource",
		"Edit a Kubernetes resource from the default editor.",
	)
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

func TestCreateDeleteCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateDeleteCommand(path) },
		"delete",
		"Delete resources",
		"Delete Kubernetes resources by file names, stdin, resources and names, or by resources and label selector.",
	)
}

func TestCreateCreateCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateCreateCommand(path) },
		"create",
		"Create resources",
		"Create Kubernetes resources from files or stdin.",
	)
}

func TestCreateDeleteCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateDeleteCommand("/path/to/kubeconfig")

	// Verify that kubectl delete flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("filename"), "expected --filename flag to be present")
	require.NotNil(t, flags.Lookup("recursive"), "expected --recursive flag to be present")
	require.NotNil(t, flags.Lookup("force"), "expected --force flag to be present")
	require.NotNil(t, flags.Lookup("grace-period"), "expected --grace-period flag to be present")
	require.NotNil(t, flags.Lookup("all"), "expected --all flag to be present")
	require.NotNil(t, flags.Lookup("selector"), "expected --selector flag to be present")
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

func TestCreateExplainCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateExplainCommand(path) },
		"explain",
		"Get documentation for a resource",
		"Get documentation for Kubernetes resources, including field descriptions and structure.",
	)
}

func TestCreateExplainCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateExplainCommand("/path/to/kubeconfig")

	// Verify that kubectl explain flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("recursive"), "expected --recursive flag to be present")
	require.NotNil(t, flags.Lookup("api-version"), "expected --api-version flag to be present")
	require.NotNil(t, flags.Lookup("output"), "expected --output flag to be present")
}
