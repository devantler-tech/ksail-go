package kindprovisioner_test

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var errBoom = errors.New("boom")

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	cases := testutils.DefaultNameCases("cfg-name")
	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		runActionSuccess(
			t,
			"Create()",
			nameCase.InputName,
			nameCase.ExpectedName,
			func(p *kindprovisioner.MockKindProvider, name string) {
				p.On("Create", name, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			func(prov *kindprovisioner.KindClusterProvisioner, name string) error {
				return prov.Create(name)
			},
		)
	})
}

func TestCreate_Error_CreateFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Create", "my-cluster", mock.Anything, mock.Anything, mock.Anything).Return(errBoom)

	// Act
	err := provisioner.Create("my-cluster")

	// Assert
	testutils.AssertErrWrappedContains(t, err, errBoom, "", "Create()")
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	// order doesn't matter for copy detection; reusing the same helper
	cases := []testutils.NameCase{
		{Name: "without name uses cfg", InputName: "", ExpectedName: "cfg-name"},
		{Name: "with name", InputName: "custom", ExpectedName: "custom"},
	}

	testutils.RunNameCases(t, cases, func(t *testing.T, nameCase testutils.NameCase) {
		t.Helper()
		runActionSuccess(
			t,
			"Delete()",
			nameCase.InputName,
			nameCase.ExpectedName,
			func(p *kindprovisioner.MockKindProvider, name string) {
				p.On("Delete", name, mock.Anything).Return(nil)
			},
			func(prov *kindprovisioner.KindClusterProvisioner, name string) error {
				return prov.Delete(name)
			},
		)
	})
}

func TestDelete_Error_DeleteFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Delete", "bad", mock.Anything).Return(errBoom)

	// Act
	err := provisioner.Delete("bad")

	// Assert
	testutils.AssertErrWrappedContains(t, err, errBoom, "", "Delete()")
}

func TestExists_Success_False(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"x", "y"}, nil)

	// Act
	exists, err := provisioner.Exists("not-here")

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
	exists, err := provisioner.Exists("")

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
	provider.On("List").Return(nil, errBoom)

	// Act
	exists, err := provisioner.Exists("any")

	// Assert
	if exists {
		t.Fatalf("Exists() got true, want false when error occurs")
	}

	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to list kind clusters", "Exists()")
}

func TestList_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"a", "b"}, nil)

	// Act
	got, err := provisioner.List()

	// Assert
	assert.NoError(t, err, "List()")
	assert.Equal(t, []string{"a", "b"}, got, "List()")
}

func TestList_Error_ListFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return(nil, errBoom)

	// Act
	_, err := provisioner.List()

	// Assert
	testutils.AssertErrWrappedContains(t, err, errBoom, "failed to list kind clusters", "List()")
}

func TestStart_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Start", func(p *kindprovisioner.KindClusterProvisioner) error {
		return p.Start("")
	})
}

func TestStart_Error_NoNodesFound(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return(nil, errBoom)

	// Act
	err := provisioner.Start("")

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
	err := provisioner.Start("")

	// Assert
	if err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
}

func TestStart_Error_DockerStartFailed(t *testing.T) {
	t.Parallel()
	runDockerOperationFailureTest(
		t,
		"Start",
		func(client *provisioner.MockContainerAPIClient) {
			client.On("ContainerStart", mock.Anything, "kind-control-plane", mock.Anything).Return(errBoom)
		},
		func(p *kindprovisioner.KindClusterProvisioner) error {
			return p.Start("")
		},
		"docker start failed for kind-control-plane",
	)
}

func TestStop_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Stop", func(p *kindprovisioner.KindClusterProvisioner) error {
		return p.Stop("")
	})
}

func TestStop_Error_NoNodesFound(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return(nil, errBoom)

	// Act
	err := provisioner.Stop("")

	// Assert
	if err == nil {
		t.Fatalf("Stop() expected error, got nil")
	}
}

func TestStop_Error_DockerStopFailed(t *testing.T) {
	t.Parallel()
	runDockerOperationFailureTest(
		t,
		"Stop",
		func(client *provisioner.MockContainerAPIClient) {
			client.On("ContainerStop", mock.Anything, "kind-control-plane", mock.Anything).Return(errBoom)
		},
		func(p *kindprovisioner.KindClusterProvisioner) error {
			return p.Stop("")
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
	err := provisioner.Stop("")

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

	cfg := &v1alpha4.Cluster{Name: "cfg-name"}
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

// helper to run a successful action (Create/Delete) flow with expectation and assertion.
type expectProviderFn func(*kindprovisioner.MockKindProvider, string)
type actionFn func(*kindprovisioner.KindClusterProvisioner, string) error

func runActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectProviderFn,
	action actionFn,
) {
	t.Helper()
	provisioner, provider, _ := newProvisionerForTest(t)
	expect(provider, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

// runDockerOperationFailureTest is a helper to DRY up the repeated Docker operation failure scenarios.
func runDockerOperationFailureTest(
	t *testing.T,
	actionName string,
	expectDockerCall func(*provisioner.MockContainerAPIClient),
	action func(*kindprovisioner.KindClusterProvisioner) error,
	expectedErrorMsg string,
) {
	t.Helper()
	// Arrange
	provisioner, provider, client := newProvisionerForTest(t)

	provider.On("ListNodes", "cfg-name").Return([]string{"kind-control-plane"}, nil)

	expectDockerCall(client)

	// Act
	err := action(provisioner)

	// Assert
	testutils.AssertErrWrappedContains(t, err, errBoom, expectedErrorMsg, actionName+"()")
}
