package kindprovisioner_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/docker"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/log"
)

// NOTE: These tests demonstrate the POC implementation works, but also
// highlight the limitations and complexity compared to the Provider-based approach.

var errPOCTest = errors.New("poc test error")

// mockKindCommandRunner captures command execution for testing.
type mockKindCommandRunner struct {
	recordedArgs []string
	stdout       string
	stderr       string
	err          error
}

func (m *mockKindCommandRunner) Run(
	_ context.Context,
	_ *cobra.Command,
	args []string,
) (stdout, stderr string, err error) {
	m.recordedArgs = args
	return m.stdout, m.stderr, m.err
}

// mockCommandBuilder captures builder calls for testing.
type mockCommandBuilder struct {
	called bool
}

func (m *mockCommandBuilder) build(logger log.Logger, streams kindcmd.IOStreams) *cobra.Command {
	m.called = true
	cmd := &cobra.Command{
		Use:  "test",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	return cmd
}

func TestPOC_ListSuccess(t *testing.T) {
	t.Parallel()

	runner := &mockKindCommandRunner{
		stdout: "cluster1\ncluster2\n",
		stderr: "",
		err:    nil,
	}

	kindConfig := &v1alpha4.Cluster{Name: "test-cluster"}
	mockProvider := kindprovisioner.NewMockKindProvider(t)
	mockClient := docker.NewMockContainerAPIClient(t)

	prov := kindprovisioner.NewKindCommandProvisionerPOC(
		kindConfig,
		"",
		mockClient,
		mockProvider,
		kindprovisioner.WithKindCommandRunner(runner),
	)

	clusters, err := prov.List(context.Background())

	require.NoError(t, err)
	assert.Equal(t, []string{"cluster1", "cluster2"}, clusters)
}

func TestPOC_ListErrorCommandFailed(t *testing.T) {
	t.Parallel()

	runner := &mockKindCommandRunner{
		stdout: "",
		stderr: "some error",
		err:    errPOCTest,
	}

	kindConfig := &v1alpha4.Cluster{Name: "test-cluster"}
	mockProvider := kindprovisioner.NewMockKindProvider(t)
	mockClient := docker.NewMockContainerAPIClient(t)

	prov := kindprovisioner.NewKindCommandProvisionerPOC(
		kindConfig,
		"",
		mockClient,
		mockProvider,
		kindprovisioner.WithKindCommandRunner(runner),
	)

	clusters, err := prov.List(context.Background())

	require.Error(t, err)
	assert.Nil(t, clusters)
	assert.Contains(t, err.Error(), "failed to list kind clusters")
	assert.Contains(t, err.Error(), "some error")
}

func TestPOC_DeleteUsesCorrectName(t *testing.T) {
	t.Parallel()

	runner := &mockKindCommandRunner{
		stdout: "",
		stderr: "",
		err:    nil,
	}

	kindConfig := &v1alpha4.Cluster{Name: "config-cluster"}
	mockProvider := kindprovisioner.NewMockKindProvider(t)
	mockClient := docker.NewMockContainerAPIClient(t)

	prov := kindprovisioner.NewKindCommandProvisionerPOC(
		kindConfig,
		"",
		mockClient,
		mockProvider,
		kindprovisioner.WithKindCommandRunner(runner),
	)

	err := prov.Delete(context.Background(), "explicit-name")

	require.NoError(t, err)
	assert.Contains(t, runner.recordedArgs, "explicit-name")
	assert.Contains(t, runner.recordedArgs, "--name")
}

func TestPOC_DeleteDefaultsToConfigName(t *testing.T) {
	t.Parallel()

	runner := &mockKindCommandRunner{
		stdout: "",
		stderr: "",
		err:    nil,
	}

	kindConfig := &v1alpha4.Cluster{Name: "config-cluster"}
	mockProvider := kindprovisioner.NewMockKindProvider(t)
	mockClient := docker.NewMockContainerAPIClient(t)

	prov := kindprovisioner.NewKindCommandProvisionerPOC(
		kindConfig,
		"",
		mockClient,
		mockProvider,
		kindprovisioner.WithKindCommandRunner(runner),
	)

	err := prov.Delete(context.Background(), "")

	require.NoError(t, err)
	assert.Contains(t, runner.recordedArgs, "config-cluster")
}

func TestPOC_ExistsReturnsTrue(t *testing.T) {
	t.Parallel()

	runner := &mockKindCommandRunner{
		stdout: "cluster1\ncluster2\nmy-cluster\n",
		stderr: "",
		err:    nil,
	}

	kindConfig := &v1alpha4.Cluster{Name: "my-cluster"}
	mockProvider := kindprovisioner.NewMockKindProvider(t)
	mockClient := docker.NewMockContainerAPIClient(t)

	prov := kindprovisioner.NewKindCommandProvisionerPOC(
		kindConfig,
		"",
		mockClient,
		mockProvider,
		kindprovisioner.WithKindCommandRunner(runner),
	)

	exists, err := prov.Exists(context.Background(), "")

	require.NoError(t, err)
	assert.True(t, exists)
}

func TestPOC_ExistsReturnsFalse(t *testing.T) {
	t.Parallel()

	runner := &mockKindCommandRunner{
		stdout: "cluster1\ncluster2\n",
		stderr: "",
		err:    nil,
	}

	kindConfig := &v1alpha4.Cluster{Name: "my-cluster"}
	mockProvider := kindprovisioner.NewMockKindProvider(t)
	mockClient := docker.NewMockContainerAPIClient(t)

	prov := kindprovisioner.NewKindCommandProvisionerPOC(
		kindConfig,
		"",
		mockClient,
		mockProvider,
		kindprovisioner.WithKindCommandRunner(runner),
	)

	exists, err := prov.Exists(context.Background(), "")

	require.NoError(t, err)
	assert.False(t, exists)
}

func TestPOC_SimpleKindRunnerBasicExecution(t *testing.T) {
	t.Parallel()

	runner := kindprovisioner.NewSimpleKindRunner()

	var outBuf, errBuf bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.OutOrStdout().Write([]byte("output"))
			cmd.ErrOrStderr().Write([]byte("error"))
			return nil
		},
	}
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	stdout, stderr, err := runner.Run(context.Background(), cmd, []string{})

	require.NoError(t, err)
	assert.Contains(t, stdout, "output")
	assert.Contains(t, stderr, "error")
}

func TestPOC_SimpleKindRunnerReturnsError(t *testing.T) {
	t.Parallel()

	runner := kindprovisioner.NewSimpleKindRunner()

	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errPOCTest
		},
	}

	_, _, err := runner.Run(context.Background(), cmd, []string{})

	require.Error(t, err)
	assert.ErrorIs(t, err, errPOCTest)
}

// TestPOC_CreateSerializesConfig demonstrates the temp file limitation.
// Note: This test shows Create will fail because it requires Docker and actual file I/O.
// In a real scenario, this complexity is a major drawback vs the Provider interface.
func TestPOC_CreateHighlightsComplexity(t *testing.T) {
	t.Parallel()

	// This test shows that Create is more complex with POC approach:
	// 1. Must handle temp file creation
	// 2. Must serialize YAML
	// 3. Must clean up files
	// 4. More failure modes than Provider interface

	kindConfig := &v1alpha4.Cluster{Name: "test-cluster"}
	mockProvider := kindprovisioner.NewMockKindProvider(t)
	mockClient := docker.NewMockContainerAPIClient(t)

	// Create will fail in test environment, but demonstrates the API
	prov := kindprovisioner.NewKindCommandProvisionerPOC(
		kindConfig,
		"",
		mockClient,
		mockProvider,
	)

	// This call will attempt to:
	// 1. Create temp file
	// 2. Marshal config to YAML
	// 3. Write temp file
	// 4. Execute kind create command
	// 5. Clean up temp file
	// Much more complex than: provider.Create(name, cluster.CreateWithV1Alpha4Config(cfg))
	err := prov.Create(context.Background(), "test")

	// We expect this to fail in test environment (no Docker, etc)
	// The point is to demonstrate the complexity, not make it work
	assert.Error(t, err)
}
