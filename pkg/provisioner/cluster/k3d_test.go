package clusterprovisioner_test

import (
	"errors"
	"testing"

	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/types"
	"github.com/stretchr/testify/mock"
)

var errK3dBoom = errors.New("k3d boom")

func TestK3dCreate_Success(t *testing.T) {
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
			runK3dActionSuccess(
				t,
				"Create()",
				testCase.inputName,
				testCase.expectedName,
				//nolint:lll // Long function signature for test readability
				func(clientProvider *clusterprovisioner.MockK3dClientProvider, configProvider *clusterprovisioner.MockK3dConfigProvider, _ string) {
					//nolint:exhaustruct // Test struct doesn't need all fields
					//nolint:lll // Long line for test readability
					configProvider.On("TransformSimpleToClusterConfig", mock.Anything, mock.Anything, mock.Anything, "k3d.yaml").Return(&v1alpha5.ClusterConfig{}, nil)
					clientProvider.On("ClusterRun", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				},
				func(prov *clusterprovisioner.K3dClusterProvisioner, name string) error {
					return prov.Create(name)
				},
			)
		})
	}
}

func TestK3dCreate_Error_TransformFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, _, configProvider := newK3dProvisionerForTest(t)
	//nolint:lll // Long line for test readability
	configProvider.On("TransformSimpleToClusterConfig", mock.Anything, mock.Anything, mock.Anything, "k3d.yaml").Return(nil, errK3dBoom)

	// Act
	err := provisioner.Create("my-cluster")

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "transform simple to cluster config", "Create()")
}

func TestK3dCreate_Error_ClusterRunFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, configProvider := newK3dProvisionerForTest(t)

	//nolint:exhaustruct // Test struct doesn't need all fields
	configProvider.On("TransformSimpleToClusterConfig", mock.Anything, mock.Anything, mock.Anything, "k3d.yaml").Return(&v1alpha5.ClusterConfig{}, nil)
	clientProvider.On("ClusterRun", mock.Anything, mock.Anything, mock.Anything).Return(errK3dBoom)

	// Act
	err := provisioner.Create("my-cluster")

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "cluster run", "Create()")
}

func TestK3dDelete_Success(t *testing.T) {
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
			runK3dActionSuccess(
				t,
				"Delete()",
				testCase.inputName,
				testCase.expectedName,
				//nolint:lll // Long function signature for test readability
				func(clientProvider *clusterprovisioner.MockK3dClientProvider, _ *clusterprovisioner.MockK3dConfigProvider, name string) {
					clientProvider.On("ClusterDelete", mock.Anything, mock.Anything, mock.MatchedBy(func(cluster *types.Cluster) bool {
						return cluster.Name == name
					}), mock.Anything).Return(nil)
				},
				func(prov *clusterprovisioner.K3dClusterProvisioner, name string) error {
					return prov.Delete(name)
				},
			)
		})
	}
}

func TestK3dDelete_Error_DeleteFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterDelete", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errK3dBoom)

	// Act
	err := provisioner.Delete("bad")

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "cluster delete", "Delete()")
}

func TestK3dStart_Success(t *testing.T) {
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
			runK3dActionSuccess(
				t,
				"Start()",
				testCase.inputName,
				testCase.expectedName,
				func(clientProvider *clusterprovisioner.MockK3dClientProvider, _ *clusterprovisioner.MockK3dConfigProvider, name string) {
					cluster := &types.Cluster{Name: name}
					clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.MatchedBy(func(c *types.Cluster) bool {
						return c.Name == name
					})).Return(cluster, nil)
					clientProvider.On("ClusterStart", mock.Anything, mock.Anything, cluster, mock.Anything).Return(nil)
				},
				func(prov *clusterprovisioner.K3dClusterProvisioner, name string) error {
					return prov.Start(name)
				},
			)
		})
	}
}

func TestK3dStart_Error_GetFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	// Act
	err := provisioner.Start("my-cluster")

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "cluster get", "Start()")
}

func TestK3dStart_Error_StartFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)

	cluster := &types.Cluster{Name: "my-cluster"}
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).Return(cluster, nil)
	clientProvider.On("ClusterStart", mock.Anything, mock.Anything, cluster, mock.Anything).Return(errK3dBoom)

	// Act
	err := provisioner.Start("my-cluster")

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "cluster start", "Start()")
}

func TestK3dStop_Success(t *testing.T) {
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
			runK3dActionSuccess(
				t,
				"Stop()",
				testCase.inputName,
				testCase.expectedName,
				//nolint:lll // Long function signature for test readability
				func(clientProvider *clusterprovisioner.MockK3dClientProvider, _ *clusterprovisioner.MockK3dConfigProvider, name string) {
					cluster := &types.Cluster{Name: name}
					clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.MatchedBy(func(c *types.Cluster) bool {
						return c.Name == name
					})).Return(cluster, nil)
					clientProvider.On("ClusterStop", mock.Anything, mock.Anything, cluster).Return(nil)
				},
				func(prov *clusterprovisioner.K3dClusterProvisioner, name string) error {
					return prov.Stop(name)
				},
			)
		})
	}
}

func TestK3dStop_Error_GetFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	// Act
	err := provisioner.Stop("my-cluster")

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "cluster get", "Stop()")
}

func TestK3dStop_Error_StopFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)

	cluster := &types.Cluster{Name: "my-cluster"}
	clientProvider.On("ClusterGet", mock.Anything, mock.Anything, mock.Anything).Return(cluster, nil)
	clientProvider.On("ClusterStop", mock.Anything, mock.Anything, cluster).Return(errK3dBoom)

	// Act
	err := provisioner.Stop("my-cluster")

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "cluster stop", "Stop()")
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
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(got) != 2 || got[0] != "cluster-a" || got[1] != "cluster-b" {
		t.Fatalf("List() got %v, want [cluster-a cluster-b]", got)
	}
}

func TestK3dList_Error_ListFailed(t *testing.T) {
	t.Parallel()
	// Arrange
	provisioner, clientProvider, _ := newK3dProvisionerForTest(t)
	clientProvider.On("ClusterList", mock.Anything, mock.Anything).Return(nil, errK3dBoom)

	// Act
	_, err := provisioner.List()

	// Assert
	assertErrWrappedContains(t, err, errK3dBoom, "cluster list", "List()")
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
	*clusterprovisioner.K3dClusterProvisioner,
	*clusterprovisioner.MockK3dClientProvider,
	*clusterprovisioner.MockK3dConfigProvider,
) {
	t.Helper()
	clientProvider := clusterprovisioner.NewMockK3dClientProvider(t)
	configProvider := clusterprovisioner.NewMockK3dConfigProvider(t)

	//nolint:exhaustruct // Test struct doesn't need all fields
	cfg := &v1alpha5.SimpleConfig{}
	cfg.Name = "cfg-name"
	provisioner := clusterprovisioner.NewK3dClusterProvisionerWithProviders(cfg, clientProvider, configProvider)

	return provisioner, clientProvider, configProvider
}

// helper to run a successful k3d action (Create/Delete/Start/Stop) flow with expectation and assertion.
//
//nolint:lll // Long line for test readability
type expectK3dProviderFn func(*clusterprovisioner.MockK3dClientProvider, *clusterprovisioner.MockK3dConfigProvider, string)
type k3dActionFn func(*clusterprovisioner.K3dClusterProvisioner, string) error

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
