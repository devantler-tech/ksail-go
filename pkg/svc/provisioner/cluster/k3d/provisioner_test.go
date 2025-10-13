package k3dprovisioner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/k3d"
	"github.com/docker/go-connections/nat"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errK3dBoom = errors.New("k3d boom")

func TestK3dCreateSuccess(t *testing.T) {
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
			clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).Return(&types.Cluster{
				Name: "cfg-name",
				Nodes: []*types.Node{
					{Name: "k3d-cfg-name-server-0", Role: types.ServerRole},
				},
			}, nil)
			clientProvider.On("KubeconfigGet", mock.Anything, mock.Anything, mock.Anything).Return(&clientcmdapi.Config{}, nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Create(context.Background(), name)
		},
	)
}

func TestK3dCreateErrorTransformFailed(t *testing.T) {
	t.Parallel()
	provisioner, _, configProvider := newK3dProvisionerForTest(t)
	expectTransformSimpleToClusterConfigErr(configProvider, errK3dBoom)

	err := provisioner.Create(context.Background(), "my-cluster")

	testutils.AssertErrWrappedContains(
		t,
		err,
		errK3dBoom,
		"transform simple to cluster config",
		"Create()",
	)
}

func TestK3dCreateErrorClusterRunFailed(t *testing.T) {
	t.Parallel()
	provisioner, clientProvider, configProvider := newK3dProvisionerForTest(t)

	expectTransformSimpleToClusterConfigOK(configProvider)
	clientProvider.On("ClusterRun", mock.Anything, mock.Anything, mock.Anything).Return(errK3dBoom)

	err := provisioner.Create(context.Background(), "my-cluster")

	testutils.AssertErrWrappedContains(t, err, errK3dBoom, "cluster run", "Create()")
}

func TestK3dDeleteSuccess(t *testing.T) {
	t.Parallel()
	runK3dNamedActionCases(
		t,
		"Delete()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, _ *k3dprovisioner.MockK3dConfigProvider, name string) {
			clientProvider.On("ClusterDelete", mock.Anything, mock.Anything, mock.MatchedBy(func(cluster *types.Cluster) bool {
				return cluster.Name == name
			}), mock.Anything).
				Return(nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Delete(context.Background(), name)
		},
	)
}

func TestK3dDeleteErrorDeleteFailed(t *testing.T) {
	t.Parallel()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterDelete", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(errK3dBoom)

	err := provisioner.Delete(context.Background(), "bad")

	testutils.AssertErrWrappedContains(t, err, errK3dBoom, "cluster delete", "Delete()")
}

func TestK3dStartSuccess(t *testing.T) {
	t.Parallel()
	runK3dNamedActionCases(
		t,
		"Start()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, _ *k3dprovisioner.MockK3dConfigProvider, name string) {
			cluster := expectClusterGetByName(clientProvider, name)
			clientProvider.On("ClusterStart", mock.Anything, mock.Anything, cluster, mock.Anything).
				Return(nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Start(context.Background(), name)
		},
	)
}

func TestK3dStartErrorGetFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterGetError(t, "Start()", func(p *k3dprovisioner.K3dClusterProvisioner) error {
		return p.Start(context.Background(), "my-cluster")
	})
}

func TestK3dStartErrorStartFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterOpErrorAfterGet(
		t,
		"Start()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, cluster *types.Cluster) {
			clientProvider.On("ClusterStart", mock.Anything, mock.Anything, cluster, mock.Anything).
				Return(errK3dBoom)
		},
		func(p *k3dprovisioner.K3dClusterProvisioner) error {
			return p.Start(context.Background(), "my-cluster")
		},
		"cluster start",
	)
}

func TestK3dStopSuccess(t *testing.T) {
	t.Parallel()
	runK3dNamedActionCases(
		t,
		"Stop()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, _ *k3dprovisioner.MockK3dConfigProvider, name string) {
			cluster := expectClusterGetByName(clientProvider, name)
			clientProvider.On("ClusterStop", mock.Anything, mock.Anything, cluster).Return(nil)
		},
		func(prov *k3dprovisioner.K3dClusterProvisioner, name string) error {
			return prov.Stop(context.Background(), name)
		},
	)
}

func TestK3dStopErrorGetFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterGetError(t, "Stop()", func(p *k3dprovisioner.K3dClusterProvisioner) error {
		return p.Stop(context.Background(), "my-cluster")
	})
}

func TestK3dStopErrorStopFailed(t *testing.T) {
	t.Parallel()
	runK3dClusterOpErrorAfterGet(
		t,
		"Stop()",
		func(clientProvider *k3dprovisioner.MockK3dClientProvider, cluster *types.Cluster) {
			clientProvider.On("ClusterStop", mock.Anything, mock.Anything, cluster).
				Return(errK3dBoom)
		},
		func(p *k3dprovisioner.K3dClusterProvisioner) error { return p.Stop(context.Background(), "my-cluster") },
		"cluster stop",
	)
}

func TestK3dListSuccess(t *testing.T) {
	t.Parallel()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clusters := []*types.Cluster{
		{Name: "cluster-a"},
		{Name: "cluster-b"},
	}
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(clusters, nil)

	got, err := provisioner.List(context.Background())

	require.NoError(t, err, "List()")
	assert.Equal(t, []string{"cluster-a", "cluster-b"}, got, "List()")
}

func TestK3dListErrorListFailed(t *testing.T) {
	t.Parallel()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	_, err := provisioner.List(context.Background())

	testutils.AssertErrWrappedContains(t, err, errK3dBoom, "cluster list", "List()")
}

func TestK3dExistsSuccessFalse(t *testing.T) {
	t.Parallel()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clusters := []*types.Cluster{
		{Name: "cluster-x"},
		{Name: "cluster-y"},
	}
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(clusters, nil)

	exists, err := provisioner.Exists(context.Background(), "not-here")
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if exists {
		t.Fatalf("Exists() got true, want false")
	}
}

func TestK3dExistsSuccessTrue(t *testing.T) {
	t.Parallel()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clusters := []*types.Cluster{
		{Name: "cluster-x"},
		{Name: "cfg-name"},
	}
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(clusters, nil)

	exists, err := provisioner.Exists(context.Background(), "")
	if err != nil {
		t.Fatalf("Exists() unexpected error: %v", err)
	}

	if !exists {
		t.Fatalf("Exists() got false, want true")
	}
}

func TestK3dExistsErrorListFailed(t *testing.T) {
	t.Parallel()
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	exists, err := provisioner.Exists(context.Background(), "any")

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

	cfg := buildTestSimpleConfig()
	provisioner := k3dprovisioner.NewK3dClusterProvisioner(cfg, clientProvider, configProvider)

	return provisioner, clientProvider, configProvider
}

func buildTestSimpleConfig() *v1alpha5.SimpleConfig {
	cfg := &v1alpha5.SimpleConfig{}
	cfg.Name = "cfg-name"

	return cfg
}

type (
	expectK3dProviderFn func(*k3dprovisioner.MockK3dClientProvider, *k3dprovisioner.MockK3dConfigProvider, string)
	k3dActionFn         func(*k3dprovisioner.K3dClusterProvisioner, string) error
)

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

	cases := testutils.DefaultNameCases("cfg-name")
	testutils.RunNameCases(t, cases, func(t *testing.T, c testutils.NameCase) {
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
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errK3dBoom)

	err := action(provisioner)
	testutils.AssertErrWrappedContains(t, err, errK3dBoom, "cluster get", label)
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
	cluster := createClusterWithKubeAPI("my-cluster")
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).
		Return(cluster, nil)
	expectOp(clientProvider, cluster)

	err := action(provisioner)
	testutils.AssertErrWrappedContains(t, err, errK3dBoom, expectedMsg, label)
}

// expectTransformSimpleToClusterConfigOK sets up a successful TransformSimpleToClusterConfig expectation.
func expectTransformSimpleToClusterConfigOK(configProvider *k3dprovisioner.MockK3dConfigProvider) {
	configProvider.On(
		"TransformSimpleToClusterConfig",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		"k3d.yaml",
	).Return(createDefaultClusterConfig(), nil)
}

// expectTransformSimpleToClusterConfigErr sets up a failing TransformSimpleToClusterConfig expectation.
func expectTransformSimpleToClusterConfigErr(
	configProvider *k3dprovisioner.MockK3dConfigProvider,
	err error,
) {
	configProvider.On(
		"TransformSimpleToClusterConfig",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		"k3d.yaml",
	).Return(nil, err)
}

// expectClusterGetByName sets up ClusterGet to return a cluster with the given name and returns the cluster.
func expectClusterGetByName(
	clientProvider *k3dprovisioner.MockK3dClientProvider,
	name string,
) *types.Cluster {
	cluster := createDefaultCluster(name)
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.MatchedBy(func(c *types.Cluster) bool {
		return c.Name == name
	})).
		Return(cluster, nil)

	return cluster
}

// createDefaultCluster creates a default types.Cluster for testing.
func createDefaultCluster(name string) *types.Cluster {
	return &types.Cluster{
		Name: name,
	}
}

// createClusterWithKubeAPI creates a cluster with KubeAPI configuration for testing.
func createClusterWithKubeAPI(name string) *types.Cluster {
	cluster := createDefaultCluster(name)
	cluster.KubeAPI = &types.ExposureOpts{
		PortMapping: nat.PortMapping{
			Port: "",
			Binding: nat.PortBinding{
				HostIP:   "",
				HostPort: "",
			},
		},
		Host: "",
	}
	cluster.ServerLoadBalancer = &types.Loadbalancer{
		Node:   nil,
		Config: nil,
	}

	return cluster
}

// createDefaultClusterConfig creates a default v1alpha5.ClusterConfig for testing.
func createDefaultClusterConfig() *v1alpha5.ClusterConfig {
	return &v1alpha5.ClusterConfig{
		Cluster: *createDefaultCluster(""),
	}
}
