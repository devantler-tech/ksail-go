package kindprovisioner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/client/docker"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// setupKindProvisioner is a helper function that creates a Kind provisioner and mock provider for testing.
// This eliminates code duplication between Create and Delete tests.
func setupKindProvisioner(
	t *testing.T,
) (*kindprovisioner.KindClusterProvisioner, *kindprovisioner.MockKindProvider) {
	t.Helper()
	provisioner, provider, _ := newProvisionerForTest(t)

	return provisioner, provider
}

func TestCreateSuccess(t *testing.T) {
	t.Parallel()
	clustertestutils.RunCreateSuccessTest(
		t,
		setupKindProvisioner,
		func(p *kindprovisioner.MockKindProvider, name string) {
			p.On("Create", name, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		},
		func(prov *kindprovisioner.KindClusterProvisioner, name string) error {
			return prov.Create(context.Background(), name)
		},
	)
}

func TestCreateErrorCreateFailed(t *testing.T) {
	t.Parallel()
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Create", "my-cluster", mock.Anything, mock.Anything, mock.Anything).
		Return(clustertestutils.ErrCreateClusterFailed)

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
	// order doesn't matter for copy detection; reusing the same helper
	cases := clustertestutils.DefaultDeleteCases()
	clustertestutils.RunStandardSuccessTest(
		t,
		cases,
		func(t *testing.T, inputName, expectedName string) {
			t.Helper()
			clustertestutils.RunActionSuccess(
				t,
				"Delete()",
				inputName,
				expectedName,
				setupKindProvisioner,
				func(p *kindprovisioner.MockKindProvider, name string) {
					p.On("Delete", name, mock.Anything).Return(nil)
				},
				func(prov *kindprovisioner.KindClusterProvisioner, name string) error {
					return prov.Delete(context.Background(), name)
				},
			)
		},
	)
}

func TestDeleteErrorDeleteFailed(t *testing.T) {
	t.Parallel()
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Delete", "bad", mock.Anything).Return(clustertestutils.ErrDeleteClusterFailed)

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
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"x", "y"}, nil)

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
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"x", "cfg-name"}, nil)

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
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return(nil, clustertestutils.ErrListClustersFailed)

	exists, err := provisioner.Exists(context.Background(), "any")

	if exists {
		t.Fatalf("Exists() got true, want false when error occurs")
	}

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list kind clusters", "Exists()")
}

func TestListSuccess(t *testing.T) {
	t.Parallel()
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"a", "b"}, nil)

	got, err := provisioner.List(context.Background())

	require.NoError(t, err, "List()")
	assert.Equal(t, []string{"a", "b"}, got, "List()")
}

func TestListErrorListFailed(t *testing.T) {
	t.Parallel()
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return(nil, clustertestutils.ErrListClustersFailed)

	_, err := provisioner.List(context.Background())

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list kind clusters", "List()")
}

func TestStartErrorClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Start", func(p *kindprovisioner.KindClusterProvisioner) error {
		return p.Start(context.Background(), "")
	})
}

func TestStartErrorNoNodesFound(t *testing.T) {
	t.Parallel()
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return(nil, clustertestutils.ErrStartClusterFailed)

	err := provisioner.Start(context.Background(), "")
	if err == nil {
		t.Fatalf("Start() expected error, got nil")
	}
}

func TestStartSuccess(t *testing.T) {
	t.Parallel()
	provisioner, provider, client := newProvisionerForTest(t)

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
	provisioner, provider, _ := newProvisionerForTest(t)
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
	provisioner, provider, client := newProvisionerForTest(t)

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
) {
	t.Helper()
	provider := kindprovisioner.NewMockKindProvider(t)
	client := docker.NewMockContainerAPIClient(t)

	cfg := &v1alpha4.Cluster{
		Name: "cfg-name",
		TypeMeta: v1alpha4.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
	}
	provisioner := kindprovisioner.NewKindClusterProvisioner(
		cfg,
		"~/.kube/config",
		provider,
		client,
	)

	return provisioner, provider, client
}

// helper to DRY up the repeated "cluster not found" error scenario for Start/Stop.
func runClusterNotFoundTest(
	t *testing.T,
	actionName string,
	action func(*kindprovisioner.KindClusterProvisioner) error,
) {
	t.Helper()
	provisioner, provider, _ := newProvisionerForTest(t)
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
	provisioner, provider, client := newProvisionerForTest(t)

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
