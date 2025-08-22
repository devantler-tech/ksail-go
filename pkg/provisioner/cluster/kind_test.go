package clusterprovisioner_test

import (
	"errors"
	"strings"
	"testing"

	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/stretchr/testify/mock"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var errBoom = errors.New("boom")

func TestCreate_Success(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		inputName    string
		expectedName string
	}{
		{name: "with name", inputName: "my-cluster", expectedName: "my-cluster"},
		{name: "without name uses cfg", inputName: "", expectedName: "cfg-name"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runActionSuccess(
				t,
				"Create()",
				testCase.inputName,
				testCase.expectedName,
				func(p *clusterprovisioner.MockKindProvider, name string) {
					p.On("Create", name, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				},
				func(prov *clusterprovisioner.KindClusterProvisioner, name string) error {
					return prov.Create(name)
				},
			)
		})
	}
}

func TestCreate_Error_CreateFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Create", "my-cluster", mock.Anything, mock.Anything, mock.Anything).Return(errBoom)

	// Act
	err := provisioner.Create("my-cluster")

	// Assert
	assertErrWrappedContains(t, err, errBoom, "", "Create()")
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		inputName    string
		expectedName string
	}{
		{name: "without name uses cfg", inputName: "", expectedName: "cfg-name"},
		{name: "with name", inputName: "custom", expectedName: "custom"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			runActionSuccess(
				t,
				"Delete()",
				testCase.inputName,
				testCase.expectedName,
				func(p *clusterprovisioner.MockKindProvider, name string) {
					p.On("Delete", name, mock.Anything).Return(nil)
				},
				func(prov *clusterprovisioner.KindClusterProvisioner, name string) error {
					return prov.Delete(name)
				},
			)
		})
	}
}

func TestDelete_Error_DeleteFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("Delete", "bad", mock.Anything).Return(errBoom)

	// Act
	err := provisioner.Delete("bad")

	// Assert
	if err == nil {
		t.Fatalf("Delete() expected error, got nil")
	}

	if !errors.Is(err, errBoom) {
		t.Fatalf("Delete() error = %v, want wrapped errBoom", err)
	}
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

	assertErrWrappedContains(t, err, errBoom, "failed to list kind clusters", "Exists()")
}

func TestList_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return([]string{"a", "b"}, nil)

	// Act
	got, err := provisioner.List()

	// Assert
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("List() got %v, want [a b]", got)
	}
}

func TestList_Error_ListFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("List").Return(nil, errBoom)

	// Act
	_, err := provisioner.List()

	// Assert
	assertErrWrappedContains(t, err, errBoom, "failed to list kind clusters", "List()")
}

func TestStart_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Start", func(p *clusterprovisioner.KindClusterProvisioner) error {
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
		func(client *clusterprovisioner.MockDockerClient) {
			client.On("ContainerStart", mock.Anything, "kind-control-plane", mock.Anything).Return(errBoom)
		},
		func(p *clusterprovisioner.KindClusterProvisioner) error {
			return p.Start("")
		},
		"docker start failed for kind-control-plane",
	)
}

func TestStop_Error_ClusterNotFound(t *testing.T) {
	t.Parallel()
	runClusterNotFoundTest(t, "Stop", func(p *clusterprovisioner.KindClusterProvisioner) error {
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
		func(client *clusterprovisioner.MockDockerClient) {
			client.On("ContainerStop", mock.Anything, "kind-control-plane", mock.Anything).Return(errBoom)
		},
		func(p *clusterprovisioner.KindClusterProvisioner) error {
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
	*clusterprovisioner.KindClusterProvisioner,
	*clusterprovisioner.MockKindProvider,
	*clusterprovisioner.MockDockerClient,
) {
	t.Helper()
	provider := clusterprovisioner.NewMockKindProvider(t)
	client := clusterprovisioner.NewMockDockerClient(t)

	cfg := &v1alpha4.Cluster{Name: "cfg-name"}
	provisioner := clusterprovisioner.NewKindClusterProvisioner(cfg, "~/.kube/config", provider, client)

	return provisioner, provider, client
}

// helper to DRY up the repeated "cluster not found" error scenario for Start/Stop.
func runClusterNotFoundTest(
	t *testing.T,
	actionName string,
	action func(*clusterprovisioner.KindClusterProvisioner) error,
) {
	t.Helper()
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.On("ListNodes", "cfg-name").Return([]string{}, nil)

	err := action(provisioner)
	if err == nil {
		t.Fatalf("%s() expected error, got nil", actionName)
	}

	if !errors.Is(err, clusterprovisioner.ErrClusterNotFound) {
		t.Fatalf("%s() error = %v, want ErrClusterNotFound", actionName, err)
	}
}

// helper to run a successful action (Create/Delete) flow with expectation and assertion.
type expectProviderFn func(*clusterprovisioner.MockKindProvider, string)
type actionFn func(*clusterprovisioner.KindClusterProvisioner, string) error

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

// assertErrWrappedContains is a small helper to verify an error exists, wraps a target error,
// and optionally contains a given substring in its message.
func assertErrWrappedContains(t *testing.T, got error, want error, contains string, ctx string) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s expected error, got nil", ctx)
	}

	if !errors.Is(got, want) {
		t.Fatalf("%s error = %v, want wrapped %v", ctx, got, want)
	}

	if contains != "" && !strings.Contains(got.Error(), contains) {
		t.Fatalf("%s error message = %q, want to contain %s", ctx, got.Error(), contains)
	}
}

// runDockerOperationFailureTest is a helper to DRY up the repeated Docker operation failure scenarios.
func runDockerOperationFailureTest(
	t *testing.T,
	actionName string,
	expectDockerCall func(*clusterprovisioner.MockDockerClient),
	action func(*clusterprovisioner.KindClusterProvisioner) error,
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
	assertErrWrappedContains(t, err, errBoom, expectedErrorMsg, actionName+"()")
}
