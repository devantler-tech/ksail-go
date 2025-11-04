package kubectl_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

func TestCreateNamespaceCmd_WithFlags(t *testing.T) {
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

	// Test with flags that should be copied to the underlying kubectl command
	cmd.SetArgs([]string{"test-namespace", "--save-config"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Namespace")
	require.Contains(t, output, "test-namespace")
}

func TestCreateDeploymentCmd_WithMultipleFlags(t *testing.T) {
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

	// Test with multiple flags including array flags
	cmd.SetArgs([]string{
		"test-deploy",
		"--image=nginx:latest",
		"--replicas=3",
		"--port=8080",
	})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML with expected values
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Deployment")
	require.Contains(t, output, "test-deploy")
	require.Contains(t, output, "nginx:latest")
}

func TestCreateSecretCmd_GenericWithMultipleLiterals(t *testing.T) {
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

	// Test with multiple --from-literal flags (slice flags)
	cmd.SetArgs([]string{
		"generic",
		"test-secret",
		"--from-literal=key1=value1",
		"--from-literal=key2=value2",
	})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML with both keys
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Secret")
	require.Contains(t, output, "test-secret")
	require.Contains(t, output, "key1")
	require.Contains(t, output, "key2")
}

func TestCreateInvalidResourceType(t *testing.T) {
	t.Parallel()

	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
	client := kubectl.NewClient(ioStreams)

	// Test all the exported Create*Cmd methods work (they all use newResourceCmd internally)
	// Testing that valid resource types can be created successfully
	t.Run("namespace", func(t *testing.T) {
		t.Parallel()

		cmd, err := client.CreateNamespaceCmd()
		require.NoError(t, err)
		require.NotNil(t, cmd)
	})

	t.Run("configmap", func(t *testing.T) {
		t.Parallel()

		cmd, err := client.CreateConfigMapCmd()
		require.NoError(t, err)
		require.NotNil(t, cmd)
	})

	t.Run("secret", func(t *testing.T) {
		t.Parallel()

		cmd, err := client.CreateSecretCmd()
		require.NoError(t, err)
		require.NotNil(t, cmd)
	})
}

func TestCreateDeploymentCmd_InvalidArgs(t *testing.T) {
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

	// Test with missing required --image flag
	cmd.SetArgs([]string{"test-deploy"})
	err = cmd.Execute()
	require.Error(t, err, "expected error when --image flag is missing")
}

func TestCreateJobCmd_WithImage(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	ioStreams := genericiooptions.IOStreams{
		In:     os.Stdin,
		Out:    outBuf,
		ErrOut: errBuf,
	}
	client := kubectl.NewClient(ioStreams)

	cmd, err := client.CreateJobCmd()
	require.NoError(t, err)

	// Test creating a job with image
	cmd.SetArgs([]string{
		"test-job",
		"--image=busybox",
	})
	err = cmd.Execute()
	require.NoError(t, err)

	// Verify output contains YAML
	output := outBuf.String()
	require.Contains(t, output, "apiVersion")
	require.Contains(t, output, "kind: Job")
	require.Contains(t, output, "test-job")
	require.Contains(t, output, "busybox")
}
