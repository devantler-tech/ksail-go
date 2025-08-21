package clusterprovisioner

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/dockerclient"
	"go.uber.org/mock/gomock"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	nodes "sigs.k8s.io/kind/pkg/cluster/nodes"
)

func newProvisionerForTest(t *testing.T) (*KindClusterProvisioner, *MockKindProvider) {
	t.Helper()
	ctrl := gomock.NewController(t)
	provider := NewMockKindProvider(ctrl)
	client := dockerclient.NewMockDockerClient(ctrl)

	cfg := &v1alpha4.Cluster{Name: "cfg-name"}
	p := NewKindClusterProvisioner(cfg, "~/.kube/config", provider, client)

	return p, provider
}

func TestCreate_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		Create("my-cluster", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	// Act
	err := p.Create("my-cluster")

	// Assert
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
}

func TestCreate_UsesConfigNameWhenEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		Create("cfg-name", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	// Act
	err := p.Create("")

	// Assert
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
}

func TestCreate_Error(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	expected := errors.New("boom")
	provider.
		EXPECT().
		Create("my-cluster", gomock.Any(), gomock.Any(), gomock.Any()).
		Return(expected)

	// Act
	err := p.Create("my-cluster")

	// Assert
	if err == nil {
		t.Fatalf("Create() expected error, got nil")
	}
}

func TestDelete_UsesConfigNameWhenEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		Delete("cfg-name", gomock.Any()).
		Return(nil)

	// Act
	err := p.Delete("")

	// Assert
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
}

func TestDelete_WithExplicitName(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		Delete("custom", gomock.Any()).
		Return(nil)

	// Act
	err := p.Delete("custom")

	// Assert
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
}

func TestDelete_Error(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	expected := errors.New("cannot delete")
	provider.
		EXPECT().
		Delete("cfg-name", gomock.Any()).
		Return(expected)

	// Act
	err := p.Delete("")

	// Assert
	if err == nil {
		t.Fatalf("Delete() expected error, got nil")
	}
}

func TestList_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		List().
		Return([]string{"a", "b"}, nil)

	// Act
	got, err := p.List()

	// Assert
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("List() got %v, want [a b]", got)
	}
}

func TestList_Error(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		List().
		Return(nil, errors.New("list failed"))

	// Act
	_, err := p.List()

	// Assert
	if err == nil {
		t.Fatalf("List() expected error, got nil")
	}
}

func TestExists_True(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		List().
		Return([]string{"x", "cfg-name"}, nil)

	// Act
	exists, err := p.Exists("")

	// Assert
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if !exists {
		t.Fatalf("Exists() got false, want true")
	}
}

func TestExists_False(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		List().
		Return([]string{"x", "y"}, nil)

	// Act
	exists, err := p.Exists("not-here")

	// Assert
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if exists {
		t.Fatalf("Exists() got true, want false")
	}
}

func TestExists_Error(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		List().
		Return(nil, errors.New("boom"))

	// Act
	_, err := p.Exists("any")

	// Assert
	if err == nil {
		t.Fatalf("Exists() expected error, got nil")
	}
}

func TestStart_ListNodesError(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return(nil, errors.New("cannot list"))

	// Act
	err := p.Start("")

	// Assert
	if err == nil {
		t.Fatalf("Start() expected error, got nil")
	}
}

func TestStart_ClusterNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return([]nodes.Node{}, nil)

	// Act
	err := p.Start("")

	// Assert
	if err == nil {
		t.Fatalf("Start() expected error, got nil")
	}

	if !errors.Is(err, ErrClusterNotFound) {
		t.Fatalf("Start() error = %v, want ErrClusterNotFound", err)
	}
}

func TestStop_ListNodesError(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return(nil, errors.New("cannot list"))

	// Act
	err := p.Stop("")

	// Assert
	if err == nil {
		t.Fatalf("Stop() expected error, got nil")
	}
}

func TestStop_ClusterNotFound(t *testing.T) {
	t.Parallel()
	// Arrange
	p, provider := newProvisionerForTest(t)
	provider.
		EXPECT().
		ListNodes("cfg-name").
		Return([]nodes.Node{}, nil)

	// Act
	err := p.Stop("")

	// Assert
	if err == nil {
		t.Fatalf("Stop() expected error, got nil")
	}

	if !errors.Is(err, ErrClusterNotFound) {
		t.Fatalf("Stop() error = %v, want ErrClusterNotFound", err)
	}
}
