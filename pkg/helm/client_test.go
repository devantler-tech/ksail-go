package helm_test

import (
	"bytes"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/helm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
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

			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}

			client := helm.NewClient(outBuf, errBuf, testCase.kubeConfigPath)

			require.NotNil(t, client, "expected client to be created")
		})
	}
}

func TestCreateInstallCommand(t *testing.T) {
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

			outBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}

			client := helm.NewClient(outBuf, errBuf, testCase.kubeConfigPath)
			cmd := client.CreateInstallCommand()

			require.NotNil(t, cmd, "expected command to be created")
			assert.Equal(t, "install [NAME] [CHART]", cmd.Use, "expected command Use")
			assert.Equal(t, "Install Helm charts", cmd.Short, "expected command Short description")
			assert.Contains(
				t,
				cmd.Long,
				"Install Helm charts to provision workloads through KSail",
				"expected command Long description",
			)
		})
	}
}

func TestCreateInstallCommandHasBasicFlags(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Verify basic install flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("create-namespace"), "expected --create-namespace flag")
	require.NotNil(t, flags.Lookup("dry-run"), "expected --dry-run flag")
	require.NotNil(t, flags.Lookup("force"), "expected --force flag")
	require.NotNil(t, flags.Lookup("generate-name"), "expected --generate-name flag")
	require.NotNil(t, flags.Lookup("name-template"), "expected --name-template flag")
	require.NotNil(t, flags.Lookup("description"), "expected --description flag")
	require.NotNil(t, flags.Lookup("devel"), "expected --devel flag")
	require.NotNil(t, flags.Lookup("dependency-update"), "expected --dependency-update flag")
	require.NotNil(
		t,
		flags.Lookup("disable-openapi-validation"),
		"expected --disable-openapi-validation flag",
	)
	require.NotNil(t, flags.Lookup("atomic"), "expected --atomic flag")
	require.NotNil(t, flags.Lookup("skip-crds"), "expected --skip-crds flag")
	require.NotNil(
		t,
		flags.Lookup("render-subchart-notes"),
		"expected --render-subchart-notes flag",
	)
}

func TestCreateInstallCommandHasTimeoutAndWaitFlags(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Verify timeout and wait flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("timeout"), "expected --timeout flag")
	require.NotNil(t, flags.Lookup("wait"), "expected --wait flag")
	require.NotNil(t, flags.Lookup("wait-for-jobs"), "expected --wait-for-jobs flag")
}

func TestCreateInstallCommandHasVersionAndRepoFlags(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Verify version and repo flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("version"), "expected --version flag")
	require.NotNil(t, flags.Lookup("repo"), "expected --repo flag")
	require.NotNil(t, flags.Lookup("username"), "expected --username flag")
	require.NotNil(t, flags.Lookup("password"), "expected --password flag")
	require.NotNil(t, flags.Lookup("cert-file"), "expected --cert-file flag")
	require.NotNil(t, flags.Lookup("key-file"), "expected --key-file flag")
	require.NotNil(
		t,
		flags.Lookup("insecure-skip-tls-verify"),
		"expected --insecure-skip-tls-verify flag",
	)
	require.NotNil(t, flags.Lookup("plain-http"), "expected --plain-http flag")
}

func TestCreateInstallCommandHasNamespaceAndMiscFlags(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Verify namespace and misc flags are present
	flags := cmd.Flags()
	require.NotNil(t, flags.Lookup("namespace"), "expected --namespace flag")
	require.NotNil(t, flags.Lookup("no-hooks"), "expected --no-hooks flag")
	require.NotNil(t, flags.Lookup("replace"), "expected --replace flag")
}

func TestCreateInstallCommandMinimumArgs(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test that command requires at least 1 argument
	cmd.SetArgs([]string{})
	err := cmd.Execute()

	require.Error(t, err, "expected error when no arguments provided")
	assert.Contains(
		t,
		err.Error(),
		"requires at least 1 arg(s)",
		"expected minimum args error message",
	)
}

func TestCreateInstallCommandWithArguments(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test that command executes with valid arguments
	cmd.SetArgs([]string{"my-release", "my-chart"})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	// The command should execute without error (though it won't actually install anything)
	require.NoError(t, err, "expected no error when valid arguments provided")
	assert.Contains(
		t,
		outBuf.String(),
		"helm install functionality",
		"expected output message",
	)
}

func TestCreateInstallCommandWithSingleArgument(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test with single argument (should work as helm supports this)
	cmd.SetArgs([]string{"my-chart"})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	require.NoError(t, err, "expected no error with single argument")
}

func TestCreateInstallCommandWithFlags(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test with various flags
	cmd.SetArgs([]string{
		"my-release",
		"my-chart",
		"--namespace", "test-ns",
		"--create-namespace",
		"--atomic",
		"--wait",
		"--timeout", "10m",
		"--version", "1.2.3",
	})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	require.NoError(t, err, "expected no error with flags")
}

func TestCreateInstallCommandGenerateName(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test with generate-name flag
	cmd.SetArgs([]string{
		"my-chart",
		"--generate-name",
	})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	require.NoError(t, err, "expected no error with --generate-name flag")
}

func TestCreateInstallCommandDryRun(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test with dry-run flag
	cmd.SetArgs([]string{
		"my-release",
		"my-chart",
		"--dry-run",
		"client",
	})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	require.NoError(t, err, "expected no error with --dry-run flag")
}

func TestCreateInstallCommandWithRepoFlags(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test with repository authentication flags
	cmd.SetArgs([]string{
		"my-release",
		"my-chart",
		"--repo", "https://charts.example.com",
		"--username", "user",
		"--password", "pass",
	})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	require.NoError(t, err, "expected no error with repo flags")
}

func TestCreateInstallCommandWithTLSFlags(t *testing.T) {
	t.Parallel()

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	client := helm.NewClient(outBuf, errBuf, "")
	cmd := client.CreateInstallCommand()

	// Test with TLS flags
	cmd.SetArgs([]string{
		"my-release",
		"my-chart",
		"--cert-file", "/path/to/cert",
		"--key-file", "/path/to/key",
		"--insecure-skip-tls-verify",
	})
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)

	err := cmd.Execute()

	require.NoError(t, err, "expected no error with TLS flags")
}

func TestNewClientWithDifferentWriters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		outWriter      *bytes.Buffer
		errWriter      *bytes.Buffer
		kubeConfigPath string
	}{
		{
			name:           "with separate buffers",
			outWriter:      &bytes.Buffer{},
			errWriter:      &bytes.Buffer{},
			kubeConfigPath: "/path/to/config",
		},
		{
			name:           "with nil-like empty path",
			outWriter:      &bytes.Buffer{},
			errWriter:      &bytes.Buffer{},
			kubeConfigPath: "",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			client := helm.NewClient(
				testCase.outWriter,
				testCase.errWriter,
				testCase.kubeConfigPath,
			)

			require.NotNil(t, client, "expected client to be created")

			// Create a command and run it to verify writers are used
			cmd := client.CreateInstallCommand()
			cmd.SetArgs([]string{"my-chart"})
			cmd.SetOut(testCase.outWriter)
			cmd.SetErr(testCase.errWriter)

			err := cmd.Execute()
			require.NoError(t, err, "expected no error")

			// Verify output was written to the provided writer
			assert.NotEmpty(t, testCase.outWriter.String(), "expected output in out writer")
		})
	}
}
