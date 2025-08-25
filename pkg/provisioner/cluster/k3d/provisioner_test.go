package k3dprovisioner_test

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutil"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/stretchr/testify/mock"
)

var errK3dBoom = errors.New("k3d boom")

func TestK3dCreate_Success(t *testing.T) {
	t.Parallel()
	runK3dNamedActionCases(
		t,
		"Create()",
		func(
			clientProvider *k3dprovisioner.MockK3dClientProvider,
			configProvider *k3dprovisioner.MockK3dConfigProvider,
			_ string,
		) {
			expectTransformSimpleToClusterConfigOK(configProvider)
			clientProvider.On("ClusterRun", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Create(name)
		},
	)
}

func TestK3dCreate_Error_TransformFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, _, configProvider := newK3dProvisionerForTest(t)
	expectTransformSimpleToClusterConfigErr(configProvider, errK3dBoom)

	// Act
	err := provisioner.Create("my-cluster")

	// Assert
	testutil.AssertErrWrappedContains(t, err, errK3dBoom, "transform simple to cluster config", "Create()")
}

func TestK3dCreate_Error_ClusterRunFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, configProvider := newK3dProvisionerForTest(t)

	expectTransformSimpleToClusterConfigOK(configProvider)
	clientProvider.On("ClusterRun", mock.Anything, mock.Anything, mock.Anything).Return(errK3dBoom)

	// Act
	err := provisioner.Create("my-cluster")

	// Assert
	testutil.AssertErrWrappedContains(t, err, errK3dBoom, "cluster run", "Create()")
}

func TestK3dDelete_Success(t *testing.T) {
	t.Parallel()
	runK3dNamedActionCases(
		t,
		"Delete()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, _ *k3dprovisioner.MockK3dConfigProvider, name string) {
			clientProvider.On("ClusterDelete", mock.Anything, mock.Anything, mock.MatchedBy(func(cluster *types.Cluster) bool {
				return cluster.Name == name
			}), mock.Anything).Return(nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Delete(name)
		},
	)
}

func TestK3dDelete_Error_DeleteFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterDelete", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errK3dBoom)

	// Act
	err := provisioner.Delete("bad")

	// Assert
	testutil.AssertErrWrappedContains(t, err, errK3dBoom, "cluster delete", "Delete()")
}

func TestK3dStart_Success(t *testing.T) {
	t.Parallel()
	runK3dNamedActionCases(
		t,
		"Start()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, _ *k3dprovisioner.MockK3dConfigProvider, name string) {
			cluster := expectClusterGetByName(clientProvider, name)
			clientProvider.On("ClusterStart", mock.Anything, mock.Anything, cluster, mock.Anything).Return(nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Start(name)
		},
	)
}

func TestK3dStart_Error_GetFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterGetError(t, "Start()", func(p *k3dprovisioner.K3dClusterProvisioner) error {
		return p.Start("my-cluster")
	})
}

func TestK3dStart_Error_StartFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterOpErrorAfterGet(
		t,
		"Start()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, cluster *types.Cluster) {
			clientProvider.On("ClusterStart", mock.Anything, mock.Anything, cluster, mock.Anything).Return(errK3dBoom)
		},
		func(p *k3dprovisioner.K3dClusterProvisioner) error { return p.Start("my-cluster") },
		"cluster start",
	)
}

func TestK3dStop_Success(t *testing.T) {
	t.Parallel()
	runK3dNamedActionCases(
		t,
		"Stop()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, _ *k3dprovisioner.MockK3dConfigProvider, name string) {
			cluster := expectClusterGetByName(clientProvider, name)
			clientProvider.On("ClusterStop", mock.Anything, mock.Anything, cluster).Return(nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Stop(name)
		},
	)
}

func TestK3dStop_Error_GetFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterGetError(t, "Stop()", func(p *k3dprovisioner.K3dClusterProvisioner) error {
		return p.Stop("my-cluster")
	})
}

func TestK3dStop_Error_StopFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterOpErrorAfterGet(
		t,
		"Stop()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, cluster *types.Cluster) {
			clientProvider.On("ClusterStop", mock.Anything, mock.Anything, cluster).Return(errK3dBoom)
		},
		func(p *k3dprovisioner.K3dClusterProvisioner) error { return p.Stop("my-cluster") },
		"cluster stop",
	)
}

func TestK3dList_Success(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clusters := []*types.Cluster{
		{Name: "cluster-a"},
		{Name: "cluster-b"},
	}
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(clusters, nil)

	// Act
	got, err := provisioner.List()

	// Assert
	testutil.AssertNoError(t, err, "List()")
	testutil.AssertStringsEqualOrder(t, got, []string{"cluster-a", "cluster-b"}, "List()")
}

func TestK3dList_Error_ListFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	// Act
	_, err := provisioner.List()

	// Assert
	testutil.AssertErrWrappedContains(t, err, errK3dBoom, "cluster list", "List()")
}

func TestK3dExists_Success_False(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clusters := []*types.Cluster{
		{Name: "cluster-x"},
		{Name: "cluster-y"},
	}
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(clusters, nil)

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

func TestK3dExists_Success_True(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clusters := []*types.Cluster{
		{Name: "cluster-x"},
		{Name: "cfg-name"},
	}
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(clusters, nil)

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

func TestK3dExists_Error_ListFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	// Act
	exists, err := provisioner.Exists("any")

	// Assert
	if exists {
		t.Fatalf("Exists() got true, want false when error occurs")
	}

	if !errors.Is(err, errK3dBoom) {
		t.Fatalf("Exists() error = %v, want wrapped errK3dBoom", err)
	}
}

// --- test helpers ---

func newK3dProvisionerForTest(
	t *testing.T,
) (
	*k3dprovisioner.K3dClusterProvisioner,
	*k3dprovisioner.MockK3dClientProvider,
	*k3dprovisioner.MockK3dConfigProvider,
) {
	t.Helper()
	clientProvider := k3dprovisioner.NewMockK3dClientProvider(t)
	configProvider := k3dprovisioner.NewMockK3dConfigProvider(t)

	cfg := &v1alpha5.SimpleConfig{}
	cfg.Name = "cfg-name"
	provisioner := k3dprovisioner.NewK3dClusterProvisioner(cfg, clientProvider, configProvider)

	return provisioner, clientProvider, configProvider
}

type expectK3dProviderFn func(*k3dprovisioner.MockK3dClientProvider, *k3dprovisioner.MockK3dConfigProvider, string)
type k3dActionFn func(*k3dprovisioner.K3dClusterProvisioner, string) error

func runK3dActionSuccess(
	t *testing.T,
	label string,
	inputName, expectedName string,
	expect expectK3dProviderFn,
	action k3dActionFn,
) {
	t.Helper()
	provisioner, clientProvider, configProvider := newK3dProvisionerForTest(t)
	expect(clientProvider, configProvider, expectedName)

	err := action(provisioner, inputName)
	if err != nil {
		t.Fatalf("%s unexpected error: %v", label, err)
	}
}

// runK3dNamedActionCases wraps the common two-case pattern for name handling
// and executes the provided expectation+action for each.
func runK3dNamedActionCases(
	t *testing.T,
	label string,
	expect expectK3dProviderFn,
	action k3dActionFn,
) {
	t.Helper()

	cases := testutil.DefaultNameCases("cfg-name")
	testutil.RunNameCases(t, cases, func(t *testing.T, c testutil.NameCase) {
		t.Helper()
		runK3dActionSuccess(t, label, c.InputName, c.ExpectedName, expect, action)
	})
}

// runK3dClusterGetError DRYs the repeated "ClusterGet" failure scenario
// for Start/Stop flows.
func runK3dClusterGetError(
	t *testing.T,
	label string,
	action func(*k3dprovisioner.K3dClusterProvisioner) error,
) {
	t.Helper()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	err := action(provisioner)
	testutil.AssertErrWrappedContains(t, err, errK3dBoom, "cluster get", label)
}

// runK3dClusterOpErrorAfterGet DRYs the scenario where ClusterGet succeeds
// but the subsequent operation (start/stop) fails with errK3dBoom.
func runK3dClusterOpErrorAfterGet(
	t *testing.T,
	label string,
	expectOp func(*k3dprovisioner.MockK3dClientProvider, *types.Cluster),
	action func(*k3dprovisioner.K3dClusterProvisioner) error,
	expectedMsg string,
) {
	t.Helper()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	cluster := &types.Cluster{Name: "my-cluster"}
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).Return(cluster, nil)
	expectOp(clientProvider, cluster)

	err := action(provisioner)
	testutil.AssertErrWrappedContains(t, err, errK3dBoom, expectedMsg, label)
}

// expectTransformSimpleToClusterConfigOK sets up a successful TransformSimpleToClusterConfig expectation.
func expectTransformSimpleToClusterConfigOK(configProvider *k3dprovisioner.MockK3dConfigProvider) {
	configProvider.On(
		"TransformSimpleToClusterConfig",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		"k3d.yaml",
	).Return(&v1alpha5.ClusterConfig{}, nil)
}

// expectTransformSimpleToClusterConfigErr sets up a failing TransformSimpleToClusterConfig expectation.
func expectTransformSimpleToClusterConfigErr(configProvider *k3dprovisioner.MockK3dConfigProvider, err error) {
	configProvider.On(
		"TransformSimpleToClusterConfig",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		"k3d.yaml",
	).Return(nil, err)
}

// expectClusterGetByName sets up ClusterGet to return a cluster with the given name and returns the cluster.
func expectClusterGetByName(clientProvider *k3dprovisioner.MockK3dClientProvider, name string) *types.Cluster {
	cluster := &types.Cluster{Name: name}
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.MatchedBy(func(c *types.Cluster) bool {
		return c.Name == name
	})).Return(cluster, nil)

	return cluster
}
