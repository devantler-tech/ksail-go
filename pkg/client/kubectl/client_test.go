package kubectl_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
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

func TestCreateConfigMapCmd(t *testing.T) {
	t.Parallel()

	client := kubectl.NewClient(createTestIOStreams())
	cmd, err := client.CreateConfigMapCmd()

	require.NoError(t, err)
	require.NotNil(t, cmd)
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

func TestCreateExecCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateExecCommand(path) },
		"exec",
		"Execute a command in a container",
		"Execute a command in a container in a pod.",
	)
}

func TestCreateExecCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateExecCommand("/path/to/kubeconfig")

	// Verify that kubectl exec flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("container"), "expected --container flag to be present")
	require.NotNil(t, flags.Lookup("stdin"), "expected --stdin flag to be present")
	require.NotNil(t, flags.Lookup("tty"), "expected --tty flag to be present")
	require.NotNil(t, flags.Lookup("quiet"), "expected --quiet flag to be present")
	require.NotNil(
		t,
		flags.Lookup("pod-running-timeout"),
		"expected --pod-running-timeout flag to be present",
	)
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

func TestCreateClusterInfoCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateClusterInfoCommand(path) },
		"info",
		"Display cluster information",
		"Display addresses of the control plane and services with label kubernetes.io/cluster-service=true.",
	)
}

func TestCreateClusterInfoCommandHasSubcommands(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateClusterInfoCommand("/path/to/kubeconfig")

	// Verify that kubectl cluster-info subcommands are present
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands, "expected cluster-info command to have subcommands")

	// Check for the dump subcommand
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Name()] = true
	}

	require.True(t, subcommandNames["dump"], "expected dump subcommand to be present")
}

func TestCreateExposeCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateExposeCommand(path) },
		"expose",
		"Expose a resource as a service",
		"Expose a resource as a new Kubernetes service.",
	)
}

func TestCreateExposeCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateExposeCommand("/path/to/kubeconfig")

	// Verify that kubectl expose flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("port"), "expected --port flag to be present")
	require.NotNil(t, flags.Lookup("protocol"), "expected --protocol flag to be present")
	require.NotNil(t, flags.Lookup("target-port"), "expected --target-port flag to be present")
	require.NotNil(t, flags.Lookup("name"), "expected --name flag to be present")
	require.NotNil(t, flags.Lookup("type"), "expected --type flag to be present")
	require.NotNil(t, flags.Lookup("external-ip"), "expected --external-ip flag to be present")
}

func TestCreateGetCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateGetCommand(path) },
		"get",
		"Get resources",
		"Display one or many Kubernetes resources from your cluster.",
	)
}

func TestCreateGetCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateGetCommand("/path/to/kubeconfig")

	// Verify that kubectl get flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("filename"), "expected --filename flag to be present")
	require.NotNil(t, flags.Lookup("recursive"), "expected --recursive flag to be present")
	require.NotNil(t, flags.Lookup("watch"), "expected --watch flag to be present")
	require.NotNil(t, flags.Lookup("output"), "expected --output flag to be present")
	require.NotNil(
		t,
		flags.Lookup("all-namespaces"),
		"expected --all-namespaces flag to be present",
	)
	require.NotNil(t, flags.Lookup("selector"), "expected --selector flag to be present")
	require.NotNil(
		t,
		flags.Lookup("field-selector"),
		"expected --field-selector flag to be present",
	)
	require.NotNil(t, flags.Lookup("show-labels"), "expected --show-labels flag to be present")
}

func TestCreateScaleCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateScaleCommand(path) },
		"scale",
		"Scale resources",
		"Set a new size for a deployment, replica set, replication controller, or stateful set.",
	)
}

func TestCreateScaleCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateScaleCommand("/path/to/kubeconfig")

	// Verify that kubectl scale flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("replicas"), "expected --replicas flag to be present")
	require.NotNil(
		t,
		flags.Lookup("current-replicas"),
		"expected --current-replicas flag to be present",
	)
	require.NotNil(
		t,
		flags.Lookup("resource-version"),
		"expected --resource-version flag to be present",
	)
	require.NotNil(t, flags.Lookup("timeout"), "expected --timeout flag to be present")
}

func TestCreateLogsCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateLogsCommand(path) },
		"logs",
		"Print container logs",
		"Print the logs for a container in a pod or specified resource. "+
			"If the pod has only one container, the container name is optional.",
	)
}

func TestCreateLogsCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateLogsCommand("/path/to/kubeconfig")

	// Verify that kubectl logs flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("follow"), "expected --follow flag to be present")
	require.NotNil(t, flags.Lookup("previous"), "expected --previous flag to be present")
	require.NotNil(t, flags.Lookup("container"), "expected --container flag to be present")
	require.NotNil(t, flags.Lookup("timestamps"), "expected --timestamps flag to be present")
	require.NotNil(t, flags.Lookup("tail"), "expected --tail flag to be present")
	require.NotNil(t, flags.Lookup("since"), "expected --since flag to be present")
}

func TestCreateRolloutCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateRolloutCommand(path) },
		"rollout",
		"Manage the rollout of a resource",
		"Manage the rollout of one or many resources.",
	)
}

func TestCreateRolloutCommandHasSubcommands(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateRolloutCommand("/path/to/kubeconfig")

	// Verify that kubectl rollout subcommands are present
	subcommands := cmd.Commands()
	require.NotEmpty(t, subcommands, "expected rollout command to have subcommands")

	// Check for rollout subcommands
	subcommandNames := make(map[string]bool)
	for _, subcmd := range subcommands {
		subcommandNames[subcmd.Name()] = true
	}

	expectedSubcommands := []string{
		"history",
		"pause",
		"restart",
		"resume",
		"status",
		"undo",
	}

	for _, expected := range expectedSubcommands {
		require.True(t, subcommandNames[expected], "expected subcommand %q to be present", expected)
	}
}

func TestCreateDescribeCommand(t *testing.T) {
	t.Parallel()

	testCommandCreation(
		t,
		func(c *kubectl.Client, path string) *cobra.Command { return c.CreateDescribeCommand(path) },
		"describe",
		"Describe resources",
		"Show details of a specific resource or group of resources.",
	)
}

func TestCreateDescribeCommandHasFlags(t *testing.T) {
	t.Parallel()

	ioStreams := createTestIOStreams()

	client := kubectl.NewClient(ioStreams)
	cmd := client.CreateDescribeCommand("/path/to/kubeconfig")

	// Verify that kubectl describe flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("filename"), "expected --filename flag to be present")
	require.NotNil(t, flags.Lookup("recursive"), "expected --recursive flag to be present")
	require.NotNil(t, flags.Lookup("selector"), "expected --selector flag to be present")
	require.NotNil(
		t,
		flags.Lookup("show-events"),
		"expected --show-events flag to be present",
	)
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
