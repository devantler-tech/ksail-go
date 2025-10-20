package kindprovisioner_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/devantler-tech/ksail-go/pkg/svc/commandrunner"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/testutils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// mockCommandRunner is a test helper that mocks the command runner.
type mockCommandRunner struct {
	mock.Mock

	lastArgs []string
}

func (m *mockCommandRunner) Run(
	_ context.Context,
	_ *cobra.Command,
	args []string,
) (commandrunner.CommandResult, error) {
	callArgs := m.Called()

	// capture last arguments for tests that need to assert CLI flags
	m.lastArgs = append([]string(nil), args...)

	result, ok := callArgs.Get(0).(commandrunner.CommandResult)
	if !ok {
		err := callArgs.Error(1)
		if err != nil {
			return commandrunner.CommandResult{}, fmt.Errorf("mock run error: %w", err)
		}

		return commandrunner.CommandResult{}, nil
	}

	err := callArgs.Error(1)
	if err != nil {
		return result, fmt.Errorf("mock run error: %w", err)
	}

	return result, nil
}

func TestCreateSuccess(t *testing.T) {
	t.Parallel()

	runProvisionerRunnerSuccessTest(
		t,
		"Create",
		func(ctx context.Context, provisioner *kindprovisioner.KindClusterProvisioner, name string) error {
			return provisioner.Create(ctx, name)
		},
	)
}

func TestCreateErrorCreateFailed(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)

	// Mock command runner to return error
	runner.On("Run").
		Return(commandrunner.CommandResult{}, clustertestutils.ErrCreateClusterFailed)

	err := provisioner.Create(context.Background(), "my-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		clustertestutils.ErrCreateClusterFailed,
		"",
		"Create()",
	)
}

func TestDeleteSuccess(t *testing.T) {
	t.Parallel()

	runProvisionerRunnerSuccessTest(
		t,
		"Delete",
		func(ctx context.Context, provisioner *kindprovisioner.KindClusterProvisioner, name string) error {
			return provisioner.Delete(ctx, name)
		},
	)
}

func TestDeleteIncludesKubeconfigFlag(t *testing.T) {
	t.Parallel()

	provisioner, _, _, runner := newProvisionerForTest(t)
	runner.On("Run").Return(commandrunner.CommandResult{}, nil)

	err := provisioner.Delete(context.Background(), "")

	require.NoError(t, err, "Delete()")
	require.Contains(t, runner.lastArgs, "--kubeconfig", "Delete() should pass kubeconfig flag")
}

func TestDeleteErrorDeleteFailed(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)

	// Mock command runner to return error
	runner.On("Run").
		Return(commandrunner.CommandResult{}, clustertestutils.ErrDeleteClusterFailed)

	err := provisioner.Delete(context.Background(), "bad")

	testutils.AssertErrWrappedContains(
		t,
		err,
		clustertestutils.ErrDeleteClusterFailed,
		"",
		"Delete()",
	)
}

func TestExistsSuccessFalse(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)

	// Mock command runner to return cluster names that don't include "not-here"
	runner.On("Run").
		Return(commandrunner.CommandResult{Stdout: "x\ny\n", Stderr: ""}, nil)

	exists, err := provisioner.Exists(context.Background(), "not-here")
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if exists {
		t.Fatalf("Exists() got true, want false")
	}
}

func TestExistsSuccessTrue(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)

	// Mock command runner to return cluster names including cfg-name
	runner.On("Run").
		Return(commandrunner.CommandResult{Stdout: "x\ncfg-name\n", Stderr: ""}, nil)

	exists, err := provisioner.Exists(context.Background(), "")
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if !exists {
		t.Fatalf("Exists() got false, want true")
	}
}

func TestExistsErrorListFailed(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)

	// Mock command runner to return error
	runner.On("Run").
		Return(commandrunner.CommandResult{}, clustertestutils.ErrListClustersFailed)

	exists, err := provisioner.Exists(context.Background(), "any")

	if exists {
		t.Fatalf("Exists() got true, want false when error occurs")
	}

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list kind clusters", "Exists()")
}

func TestListSuccess(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)
	// Mock command runner to return cluster names
	runner.On("Run").
		Return(commandrunner.CommandResult{Stdout: "a\nb\n", Stderr: ""}, nil)

	got, err := provisioner.List(context.Background())

	require.NoError(t, err, "List()")
	assert.Equal(t, []string{"a", "b"}, got, "List()")
}

func TestListErrorListFailed(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)
	// Mock command runner to return error
	runner.On("Run").
		Return(commandrunner.CommandResult{}, clustertestutils.ErrListClustersFailed)

	_, err := provisioner.List(context.Background())

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list kind clusters", "List()")
}

func TestListFiltersNoKindClustersMessage(t *testing.T) {
	t.Parallel()
	provisioner, _, _, runner := newProvisionerForTest(t)

	runner.On("Run").Return(commandrunner.CommandResult{
		Stdout: "No kind clusters found.\n",
		Stderr: "",
	}, nil)

	got, err := provisioner.List(context.Background())

	require.NoError(t, err, "List()")
	require.Empty(t, got, "List() should ignore 'No kind clusters found.' message")
}

func TestStartErrorClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Start", func(p *kindprovisioner.KindClusterProvisioner) error {
		return p.Start(context.Background(), "")
	})
}

func TestStartErrorNoNodesFound(t *testing.T) {
	t.Parallel()
	provisioner, provider, _, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return(nil, clustertestutils.ErrStartClusterFailed)

	err := provisioner.Start(context.Background(), "")
	if err == nil {
		t.Fatalf("Start() expected error, got nil")
	}
}

func TestStartSuccess(t *testing.T) {
	t.Parallel()
	provisioner, provider, client, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return([]string{"kind-control-plane", "kind-worker"}, nil)

	// Expect ContainerStart called twice with any args
	client.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(2)

	err := provisioner.Start(context.Background(), "")
	if err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
}

func TestStartErrorDockerStartFailed(t *testing.T) {
	t.Parallel()
	runDockerOperationFailureTest(
		t,
		func(p *kindprovisioner.KindClusterProvisioner) error { return p.Start(context.Background(), "") },
		"Start",
		func(client *docker.MockContainerAPIClient) {
			client.On("ContainerStart", mock.Anything, "kind-control-plane", mock.Anything).
				Return(clustertestutils.ErrStartClusterFailed)
		},
		"docker start failed for kind-control-plane",
	)
}

func TestStopErrorClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Stop", func(p *kindprovisioner.KindClusterProvisioner) error {
		return p.Stop(context.Background(), "")
	})
}

func TestStopErrorNoNodesFound(t *testing.T) {
	t.Parallel()
	provisioner, provider, _, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return(nil, clustertestutils.ErrStopClusterFailed)

	err := provisioner.Stop(context.Background(), "")
	if err == nil {
		t.Fatalf("Stop() expected error, got nil")
	}
}

func TestStopErrorDockerStopFailed(t *testing.T) {
	t.Parallel()
	runDockerOperationFailureTest(
		t,
		func(p *kindprovisioner.KindClusterProvisioner) error { return p.Stop(context.Background(), "") },
		"Stop",
		func(client *docker.MockContainerAPIClient) {
			client.On("ContainerStop", mock.Anything, "kind-control-plane", mock.Anything).
				Return(clustertestutils.ErrStopClusterFailed)
		},
		"docker stop failed for kind-control-plane",
	)
}

func TestStopSuccess(t *testing.T) {
	t.Parallel()
	provisioner, provider, client, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").
		Return([]string{"kind-control-plane", "kind-worker", "kind-worker2"}, nil)

	client.On("ContainerStop", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(3)

	err := provisioner.Stop(context.Background(), "")
	if err != nil {
		t.Fatalf("Stop() unexpected error: %v", err)
	}
}

// --- internals ---

func newProvisionerForTest(
	t *testing.T,
) (
	*kindprovisioner.KindClusterProvisioner,
	*kindprovisioner.MockKindProvider,
	*docker.MockContainerAPIClient,
	*mockCommandRunner,
) {
	t.Helper()
	provider := kindprovisioner.NewMockKindProvider(t)
	client := docker.NewMockContainerAPIClient(t)
	runner := &mockCommandRunner{}

	cfg := &v1alpha4.Cluster{
		Name: "cfg-name",
		TypeMeta: v1alpha4.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
	}
	provisioner := kindprovisioner.NewKindClusterProvisionerWithRunner(
		cfg,
		"~/.kube/config",
		provider,
		client,
		runner,
	)

	return provisioner, provider, client, runner
}

// helper to DRY up the repeated "cluster not found" error scenario for Start/Stop.
func runClusterNotFoundTest(
	t *testing.T,
	actionName string,
	action func(*kindprovisioner.KindClusterProvisioner) error,
) {
	t.Helper()
	provisioner, provider, _, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return([]string{}, nil)

	err := action(provisioner)
	if err == nil {
		t.Fatalf("%s() expected error, got nil", actionName)
	}

	if !errors.Is(err, kindprovisioner.ErrClusterNotFound) {
		t.Fatalf("%s() error = %v, want ErrClusterNotFound", actionName, err)
	}
}

// runDockerOperationFailureTest is a helper for testing Docker operation failures.
func runDockerOperationFailureTest(
	t *testing.T,
	operation func(*kindprovisioner.KindClusterProvisioner) error,
	operationName string,
	expectDockerCall func(*docker.MockContainerAPIClient),
	expectedErrorMsg string,
) {
	t.Helper()
	provisioner, provider, client, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return([]string{"kind-control-plane"}, nil)

	expectDockerCall(client)

	err := operation(provisioner)
	if err == nil {
		t.Fatalf("%s() expected error, got nil", operationName)
	}

	if expectedErrorMsg != "" && !assert.Contains(t, err.Error(), expectedErrorMsg) {
		t.Fatalf("%s() error should contain %q, got: %v", operationName, expectedErrorMsg, err)
	}
}

func runProvisionerRunnerSuccessTest(
	t *testing.T,
	actionName string,
	action func(context.Context, *kindprovisioner.KindClusterProvisioner, string) error,
) {
	t.Helper()

	testCases := []struct {
		name      string
		inputName string
	}{
		{
			name:      "without_name_uses_cfg",
			inputName: "",
		},
		{
			name:      "with_name",
			inputName: "my-cluster",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			provisioner, _, _, runner := newProvisionerForTest(t)
			runner.On("Run").Return(commandrunner.CommandResult{}, nil)

			err := action(context.Background(), provisioner, testCase.inputName)
			require.NoErrorf(t, err, "%s()", actionName)
		})
	}
}
