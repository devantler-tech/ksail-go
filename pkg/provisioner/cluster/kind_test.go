package clusterprovisioner_test

import (
	"errors"
	"testing"

	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"go.uber.org/mock/gomock"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var errBoom = errors.New("boom")

func TestCreate_Success_WithName(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.
		EXPECT().
		Create("my-cluster", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	// Act
	err := provisioner.Create("my-cluster")

	// Assert
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
}

func TestCreate_Success_WithoutName(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.
		EXPECT().
		Create("cfg-name", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	// Act
	err := provisioner.Create("")

	// Assert
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
}

func TestDelete_Success_WithoutName(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.
		EXPECT().
		Delete("cfg-name", gomock.Any()).
		Return(nil)

	// Act
	err := provisioner.Delete("")

	// Assert
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
}

func TestDelete_Success_WithName(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.
		EXPECT().
		Delete("custom", gomock.Any()).
		Return(nil)

	// Act
	err := provisioner.Delete("custom")

	// Assert
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
}

func TestExists_Success_False(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.
		EXPECT().
		List().
		Return([]string{"x", "y"}, nil)

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
	provider.
		EXPECT().
		List().
		Return([]string{"x", "cfg-name"}, nil)

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

func TestList_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, _ := newProvisionerForTest(t)
	provider.
		EXPECT().
		List().
		Return([]string{"a", "b"}, nil)

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
	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return(nil, errBoom)

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

	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return([]string{"kind-control-plane", "kind-worker"}, nil)

	// Expect ContainerStart called twice with any args
	client.
		EXPECT().
		ContainerStart(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(2).
		Return(nil)

	// Act
	err := provisioner.Start("")

	// Assert
	if err != nil {
		t.Fatalf("Start() unexpected error: %v", err)
	}
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
	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return(nil, errBoom)

	// Act
	err := provisioner.Stop("")

	// Assert
	if err == nil {
		t.Fatalf("Stop() expected error, got nil")
	}
}

func TestStop_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, provider, client := newProvisionerForTest(t)

	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return([]string{"kind-control-plane", "kind-worker", "kind-worker2"}, nil)

	client.
		EXPECT().
		ContainerStop(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(3).
		Return(nil)

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
	ctrl := gomock.NewController(t)
	provider := clusterprovisioner.NewMockKindProvider(ctrl)
	client := clusterprovisioner.NewMockDockerClient(ctrl)

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
	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return([]string{}, nil)

	err := action(provisioner)
	if err == nil {
		t.Fatalf("%s() expected error, got nil", actionName)
	}

	if !errors.Is(err, clusterprovisioner.ErrClusterNotFound) {
		t.Fatalf("%s() error = %v, want ErrClusterNotFound", actionName, err)
	}
}
