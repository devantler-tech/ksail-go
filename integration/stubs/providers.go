package stubs

import (
	"context"
	"errors"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/weaveworks/eksctl/pkg/actions/cluster"
	eksapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/eks"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
)

// KindProviderStub is a stub implementation of KindProvider interface.
type KindProviderStub struct {
	CreateError    error
	DeleteError    error
	ListResult     []string
	ListError      error
	ListNodesResult []string
	ListNodesError error
	
	CreateCalls    []string
	DeleteCalls    []string
	ListCalls      int
	ListNodesCalls []string
}

// NewKindProviderStub creates a new KindProviderStub with default behavior.
func NewKindProviderStub() *KindProviderStub {
	return &KindProviderStub{
		ListResult:      []string{"kind-cluster"},
		ListNodesResult: []string{"kind-cluster-control-plane"},
	}
}

// Create simulates cluster creation.
func (k *KindProviderStub) Create(name string, opts ...kindcluster.CreateOption) error {
	k.CreateCalls = append(k.CreateCalls, name)
	return k.CreateError
}

// Delete simulates cluster deletion.
func (k *KindProviderStub) Delete(name, kubeconfigPath string) error {
	k.DeleteCalls = append(k.DeleteCalls, name)
	return k.DeleteError
}

// List simulates cluster listing.
func (k *KindProviderStub) List() ([]string, error) {
	k.ListCalls++
	return k.ListResult, k.ListError
}

// ListNodes simulates node listing for a cluster.
func (k *KindProviderStub) ListNodes(name string) ([]string, error) {
	k.ListNodesCalls = append(k.ListNodesCalls, name)
	return k.ListNodesResult, k.ListNodesError
}

// WithCreateError configures the stub to return an error on Create.
func (k *KindProviderStub) WithCreateError(message string) *KindProviderStub {
	k.CreateError = errors.New(message)
	return k
}

// WithListResult configures the stub to return specific clusters on List.
func (k *KindProviderStub) WithListResult(clusters []string) *KindProviderStub {
	k.ListResult = clusters
	return k
}

// DockerClientStub is a stub implementation of Docker container API client.
type DockerClientStub struct {
	ContainerStartError error
	ContainerStopError  error
	
	StartCalls []string
	StopCalls  []string
}

// NewDockerClientStub creates a new DockerClientStub.
func NewDockerClientStub() *DockerClientStub {
	return &DockerClientStub{}
}

// ContainerStart simulates container start.
func (d *DockerClientStub) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	d.StartCalls = append(d.StartCalls, containerID)
	return d.ContainerStartError
}

// ContainerStop simulates container stop.
func (d *DockerClientStub) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	d.StopCalls = append(d.StopCalls, containerID)
	return d.ContainerStopError
}

// K3dClientProviderStub is a stub implementation of K3dClientProvider interface.
type K3dClientProviderStub struct {
	ClusterRunError    error
	ClusterDeleteError error
	ClusterGetResult   *types.Cluster
	ClusterGetError    error
	ClusterStartError  error
	ClusterStopError   error
	ClusterListResult  []*types.Cluster
	ClusterListError   error
	
	RunCalls    int
	DeleteCalls int
	GetCalls    int
	StartCalls  int
	StopCalls   int
	ListCalls   int
}

// NewK3dClientProviderStub creates a new K3dClientProviderStub.
func NewK3dClientProviderStub() *K3dClientProviderStub {
	return &K3dClientProviderStub{
		ClusterGetResult: &types.Cluster{Name: "test-cluster"},
		ClusterListResult: []*types.Cluster{
			{Name: "k3d-cluster"},
		},
	}
}

// ClusterRun simulates cluster creation.
func (k *K3dClientProviderStub) ClusterRun(ctx context.Context, runtime runtimes.Runtime, clusterConfig *v1alpha5.ClusterConfig) error {
	k.RunCalls++
	return k.ClusterRunError
}

// ClusterDelete simulates cluster deletion.
func (k *K3dClientProviderStub) ClusterDelete(ctx context.Context, runtime runtimes.Runtime, cluster *types.Cluster, opts types.ClusterDeleteOpts) error {
	k.DeleteCalls++
	return k.ClusterDeleteError
}

// ClusterGet simulates cluster retrieval.
func (k *K3dClientProviderStub) ClusterGet(ctx context.Context, runtime runtimes.Runtime, cluster *types.Cluster) (*types.Cluster, error) {
	k.GetCalls++
	return k.ClusterGetResult, k.ClusterGetError
}

// ClusterStart simulates cluster start.
func (k *K3dClientProviderStub) ClusterStart(ctx context.Context, runtime runtimes.Runtime, cluster *types.Cluster, opts types.ClusterStartOpts) error {
	k.StartCalls++
	return k.ClusterStartError
}

// ClusterStop simulates cluster stop.
func (k *K3dClientProviderStub) ClusterStop(ctx context.Context, runtime runtimes.Runtime, cluster *types.Cluster) error {
	k.StopCalls++
	return k.ClusterStopError
}

// ClusterList simulates cluster listing.
func (k *K3dClientProviderStub) ClusterList(ctx context.Context, runtime runtimes.Runtime) ([]*types.Cluster, error) {
	k.ListCalls++
	return k.ClusterListResult, k.ClusterListError
}

// K3dConfigProviderStub is a stub implementation of K3dConfigProvider interface.
type K3dConfigProviderStub struct {
	TransformResult *v1alpha5.ClusterConfig
	TransformError  error
	
	TransformCalls int
}

// NewK3dConfigProviderStub creates a new K3dConfigProviderStub.
func NewK3dConfigProviderStub() *K3dConfigProviderStub {
	return &K3dConfigProviderStub{
		TransformResult: &v1alpha5.ClusterConfig{
			// Create minimal config structure for stub
		},
	}
}

// TransformSimpleToClusterConfig simulates config transformation.
func (k *K3dConfigProviderStub) TransformSimpleToClusterConfig(ctx context.Context, runtime runtimes.Runtime, simpleConfig v1alpha5.SimpleConfig, filename string) (*v1alpha5.ClusterConfig, error) {
	k.TransformCalls++
	return k.TransformResult, k.TransformError
}

// EKSClusterActionsStub is a stub implementation of EKSClusterActions interface.
type EKSClusterActionsStub struct {
	DeleteError error
	
	DeleteCalls int
}

// NewEKSClusterActionsStub creates a new EKSClusterActionsStub.
func NewEKSClusterActionsStub() *EKSClusterActionsStub {
	return &EKSClusterActionsStub{}
}

// Delete simulates EKS cluster deletion.
func (e *EKSClusterActionsStub) Delete(ctx context.Context, waitInterval, podEvictionWaitPeriod time.Duration, wait, force, disableNodegroupEviction bool, parallel int) error {
	e.DeleteCalls++
	return e.DeleteError
}

// EKSClusterListerStub is a stub implementation of EKSClusterLister interface.
type EKSClusterListerStub struct {
	GetClustersResult []cluster.Description
	GetClustersError  error
	
	GetClustersCalls int
}

// NewEKSClusterListerStub creates a new EKSClusterListerStub.
func NewEKSClusterListerStub() *EKSClusterListerStub {
	return &EKSClusterListerStub{
		GetClustersResult: []cluster.Description{
			{Name: "eks-cluster"},
		},
	}
}

// GetClusters simulates EKS cluster listing.
func (e *EKSClusterListerStub) GetClusters(ctx context.Context, provider *eks.ClusterProvider, listAllRegions bool, chunkSize int) ([]cluster.Description, error) {
	e.GetClustersCalls++
	return e.GetClustersResult, e.GetClustersError
}

// EKSClusterCreatorStub is a stub implementation of EKSClusterCreator interface.
type EKSClusterCreatorStub struct {
	CreateClusterError error
	
	CreateClusterCalls int
}

// NewEKSClusterCreatorStub creates a new EKSClusterCreatorStub.
func NewEKSClusterCreatorStub() *EKSClusterCreatorStub {
	return &EKSClusterCreatorStub{}
}

// CreateCluster simulates EKS cluster creation.
func (e *EKSClusterCreatorStub) CreateCluster(ctx context.Context, cfg *eksapi.ClusterConfig, ctl *eks.ClusterProvider) error {
	e.CreateClusterCalls++
	return e.CreateClusterError
}

// EKSNodeGroupManagerStub is a stub implementation of EKSNodeGroupManager interface.
type EKSNodeGroupManagerStub struct {
	ScaleError error
	
	ScaleCalls int
}

// NewEKSNodeGroupManagerStub creates a new EKSNodeGroupManagerStub.
func NewEKSNodeGroupManagerStub() *EKSNodeGroupManagerStub {
	return &EKSNodeGroupManagerStub{}
}

// Scale simulates node group scaling.
func (e *EKSNodeGroupManagerStub) Scale(ctx context.Context, ng *eksapi.NodeGroupBase, wait bool) error {
	e.ScaleCalls++
	return e.ScaleError
}

// HelmClientStub is a stub implementation of HelmClient interface.
type HelmClientStub struct {
	InstallError   error
	UninstallError error
	
	InstallCalls   []string
	UninstallCalls []string
}

// NewHelmClientStub creates a new HelmClientStub.
func NewHelmClientStub() *HelmClientStub {
	return &HelmClientStub{}
}

// Install simulates Helm chart installation.
func (h *HelmClientStub) Install(ctx context.Context, spec *helmclient.ChartSpec) error {
	h.InstallCalls = append(h.InstallCalls, spec.ReleaseName)
	return h.InstallError
}

// Uninstall simulates Helm chart uninstallation.
func (h *HelmClientStub) Uninstall(name string) error {
	h.UninstallCalls = append(h.UninstallCalls, name)
	return h.UninstallError
}