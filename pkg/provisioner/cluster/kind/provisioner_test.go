package kindprovisioner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)



func TestCreate_Success(t *testing.T) {
	t.Parallel()
	clustertestutils.RunCreateTest(t, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		clustertestutils.RunActionSuccess(
			t,
			"Create()",
			inputName,
			expectedName,
			func(t *testing.T) (*kindprovisioner.KindClusterProvisioner, *kindprovisioner.MockKindProvider) {
				t.Helper()
				provisioner, provider, _ := newProvisionerForTest(t)

				return provisioner, provider
			},
			func(p *kindprovisioner.MockKindProvider, name string) {
				p.On("Create", name, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			func(prov *kindprovisioner.KindClusterProvisioner, name string) error {
				return prov.Create(context.Background(), name)
			},
		)
	})
}

func TestCreate_Error_CreateFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Create", "my-cluster", mock.Anything, mock.Anything, mock.Anything).
		Return(clustertestutils.ErrCreateClusterFailed)

	// Act
	err := provisioner.Create(context.Background(), "my-cluster")

	// Assert
	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrCreateClusterFailed, "", "Create()")
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()
	// order doesn't matter for copy detection; reusing the same helper
	cases := clustertestutils.DefaultDeleteCases()
	clustertestutils.RunStandardSuccessTest(t, cases, func(t *testing.T, inputName, expectedName string) {
		t.Helper()
		clustertestutils.RunActionSuccess(
			t,
			"Delete()",
			inputName,
			expectedName,
			func(t *testing.T) (*kindprovisioner.KindClusterProvisioner, *kindprovisioner.MockKindProvider) {
				t.Helper()
				provisioner, provider, _ := newProvisionerForTest(t)

				return provisioner, provider
			},
			func(p *kindprovisioner.MockKindProvider, name string) {
				p.On("Delete", name, mock.Anything).Return(nil)
			},
			func(prov *kindprovisioner.KindClusterProvisioner, name string) error {
				return prov.Delete(context.Background(), name)
			},
		)
	})
}

func TestDelete_Error_DeleteFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Delete", "bad", mock.Anything).Return(clustertestutils.ErrDeleteClusterFailed)

	// Act
	err := provisioner.Delete(context.Background(), "bad")

	// Assert
	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrDeleteClusterFailed, "", "Delete()")
}

func TestExists_Success_False(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"x", "y"}, nil)

	// Act
	exists, err := provisioner.Exists(context.Background(), "not-here")

	// Assert
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if exists {
		t.Fatalf("Exists() got true, want false")
	}
}

func TestExists_Success_True(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"x", "cfg-name"}, nil)

	// Act
	exists, err := provisioner.Exists(context.Background(), "")

	// Assert
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if !exists {
		t.Fatalf("Exists() got false, want true")
	}
}

func TestExists_Error_ListFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return(nil, clustertestutils.ErrListClustersFailed)

	// Act
	exists, err := provisioner.Exists(context.Background(), "any")

	// Assert
	if exists {
		t.Fatalf("Exists() got true, want false when error occurs")
	}

	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list kind clusters", "Exists()")
}

func TestList_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"a", "b"}, nil)

	// Act
	got, err := provisioner.List(context.Background())

	// Assert
	require.NoError(t, err, "List()")
	assert.Equal(t, []string{"a", "b"}, got, "List()")
}

func TestList_Error_ListFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return(nil, clustertestutils.ErrListClustersFailed)

	// Act
	_, err := provisioner.List(context.Background())

	// Assert
	testutils.AssertErrWrappedContains(t, err, clustertestutils.ErrListClustersFailed,
		"failed to list kind clusters", "List()")
}

func TestStart_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Start", func(p *kindprovisioner.KindClusterProvisioner) error {
		return p.Start(context.Background(), "")
	})
}

func TestStart_Error_NoNodesFound(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return(nil, clustertestutils.ErrStartClusterFailed)

	// Act
	err := provisioner.Start(context.Background(), "")

	// Assert
	if err == nil {
		t.Fatalf("Start() expected error, got nil")
	}
}

func TestStart_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, client := newProvisionerForTest(t)

	provider.On("ListNodes", "cfg-name").Return([]string{"kind-control-plane", "kind-worker"}, nil)

	// Expect ContainerStart called twice with any args
	client.On("ContainerStart", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(2)

	// Act
	err := provisioner.Start(context.Background(), "")

	// Assert
	if err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
}

func TestStart_Error_DockerStartFailed(t *testing.T) {
	t.Parallel()
	runDockerOperationFailureTest(
		t,
		func(p *kindprovisioner.KindClusterProvisioner) error { return p.Start(context.Background(), "") },
		"Start",
		func(client *provisioner.MockContainerAPIClient) {
			client.On("ContainerStart", mock.Anything, "kind-control-plane", mock.Anything).
				Return(clustertestutils.ErrStartClusterFailed)
		},
		"docker start failed for kind-control-plane",
	)
}

func TestStop_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Stop", func(p *kindprovisioner.KindClusterProvisioner) error {
		return p.Stop(context.Background(), "")
	})
}

func TestStop_Error_NoNodesFound(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return(nil, clustertestutils.ErrStopClusterFailed)

	// Act
	err := provisioner.Stop(context.Background(), "")

	// Assert
	if err == nil {
		t.Fatalf("Stop() expected error, got nil")
	}
}

func TestStop_Error_DockerStopFailed(t *testing.T) {
	t.Parallel()
	runDockerOperationFailureTest(
		t,
		func(p *kindprovisioner.KindClusterProvisioner) error { return p.Stop(context.Background(), "") },
		"Stop",
		func(client *provisioner.MockContainerAPIClient) {
			client.On("ContainerStop", mock.Anything, "kind-control-plane", mock.Anything).
				Return(clustertestutils.ErrStopClusterFailed)
		},
		"docker stop failed for kind-control-plane",
	)
}

func TestStop_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, client := newProvisionerForTest(t)

	provider.On("ListNodes", "cfg-name").Return([]string{"kind-control-plane", "kind-worker", "kind-worker2"}, nil)

	client.On("ContainerStop", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(3)

	// Act
	err := provisioner.Stop(context.Background(), "")

	// Assert
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
	*provisioner.MockContainerAPIClient,
) {
	t.Helper()
	provider := kindprovisioner.NewMockKindProvider(t)
	client := provisioner.NewMockContainerAPIClient(t)

	cfg := &v1alpha4.Cluster{
		Name: "cfg-name",
		TypeMeta: v1alpha4.TypeMeta{
			Kind:       "Cluster",
			APIVersion: "kind.x-k8s.io/v1alpha4",
		},
		Nodes: []v1alpha4.Node{},
		Networking: v1alpha4.Networking{
			IPFamily:          "",
			APIServerPort:     0,
			APIServerAddress:  "",
			PodSubnet:         "",
			ServiceSubnet:     "",
			DisableDefaultCNI: false,
			KubeProxyMode:     "",
			DNSSearch:         nil,
		},
		FeatureGates:                    map[string]bool{},
		RuntimeConfig:                   map[string]string{},
		KubeadmConfigPatches:            []string{},
		KubeadmConfigPatchesJSON6902:    []v1alpha4.PatchJSON6902{},
		ContainerdConfigPatches:         []string{},
		ContainerdConfigPatchesJSON6902: []string{},
	}
	provisioner := kindprovisioner.NewKindClusterProvisioner(cfg, "~/.kube/config", provider, client)

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
	expectDockerCall func(*provisioner.MockContainerAPIClient),
	expectedErrorMsg string,
) {
	t.Helper()
	// Arrange
	provisioner, provider, client := newProvisionerForTest(t)

	provider.On("ListNodes", "cfg-name").Return([]string{"kind-control-plane"}, nil)

	expectDockerCall(client)

	// Act
	err := operation(provisioner)

	// Assert
	if err == nil {
		t.Fatalf("%s() expected error, got nil", operationName)
	}

	if expectedErrorMsg != "" && !assert.Contains(t, err.Error(), expectedErrorMsg) {
		t.Fatalf("%s() error should contain %q, got: %v", operationName, expectedErrorMsg, err)
	}
}
