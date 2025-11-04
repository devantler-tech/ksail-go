package kubectl_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestClient_CreateNamespaceCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateNamespaceCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "namespace", cmd.Name())
}

func TestClient_CreateConfigMapCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateConfigMapCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "configmap", cmd.Name())
}

func TestClient_CreateSecretCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateSecretCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "secret", cmd.Name())
	require.NotEmpty(t, cmd.Commands(), "secret command should have subcommands")
}

func TestClient_CreateServiceAccountCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateServiceAccountCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "serviceaccount", cmd.Name())
}

func TestClient_CreateDeploymentCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateDeploymentCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "deployment", cmd.Name())
}

func TestClient_CreateJobCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateJobCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "job", cmd.Name())
}

func TestClient_CreateCronJobCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateCronJobCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "cronjob", cmd.Name())
}

func TestClient_CreateServiceCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateServiceCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "service", cmd.Name())
	require.NotEmpty(t, cmd.Commands(), "service command should have subcommands")
}

func TestClient_CreateIngressCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateIngressCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "ingress", cmd.Name())
}

func TestClient_CreateRoleCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateRoleCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "role", cmd.Name())
}

func TestClient_CreateRoleBindingCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateRoleBindingCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "rolebinding", cmd.Name())
}

func TestClient_CreateClusterRoleCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateClusterRoleCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "clusterrole", cmd.Name())
}

func TestClient_CreateClusterRoleBindingCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateClusterRoleBindingCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "clusterrolebinding", cmd.Name())
}

func TestClient_CreateQuotaCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateQuotaCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "quota", cmd.Name())
}

func TestClient_CreatePodDisruptionBudgetCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreatePodDisruptionBudgetCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "poddisruptionbudget", cmd.Name())
}

func TestClient_CreatePriorityClassCmd(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreatePriorityClassCmd()
	require.NoError(t, err)
	require.NotNil(t, cmd)
	require.Equal(t, "priorityclass", cmd.Name())
}

func TestClient_ExecuteResourceGen_Namespace(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    outBuf,
		ErrOut: errBuf,
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateNamespaceCmd()
	require.NoError(t, err)

	// Execute with a namespace name
	cmd.SetArgs([]string{"test-namespace"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Namespace")
	require.Contains(t, output, "test-namespace")
}

func TestClient_ExecuteResourceGen_Deployment(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    outBuf,
		ErrOut: errBuf,
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateDeploymentCmd()
	require.NoError(t, err)

	// Execute with deployment name and image
	cmd.SetArgs([]string{"test-deploy", "--image=nginx:latest"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Deployment")
	require.Contains(t, output, "test-deploy")
	require.Contains(t, output, "nginx:latest")
}

func TestClient_ExecuteSubcommandGen_SecretGeneric(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    outBuf,
		ErrOut: errBuf,
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateSecretCmd()
	require.NoError(t, err)

	// Execute parent command with subcommand args
	cmd.SetArgs([]string{"generic", "test-secret", "--from-literal=key1=value1"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Secret")
	require.Contains(t, output, "test-secret")
}

func TestClient_ExecuteSubcommandGen_ServiceClusterIP(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    outBuf,
		ErrOut: errBuf,
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateServiceCmd()
	require.NoError(t, err)

	// Execute parent command with subcommand args
	cmd.SetArgs([]string{"clusterip", "test-service", "--tcp=8080:80"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Service")
	require.Contains(t, output, "test-service")
}
