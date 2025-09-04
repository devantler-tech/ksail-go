// Package eksprovisioner provides implementations of the Provisioner interface
// for provisioning EKS clusters in AWS.
package eksprovisioner

import (
	"context"
	"time"

	"github.com/weaveworks/eksctl/pkg/actions/cluster"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/eks"
	"k8s.io/client-go/kubernetes"
)

// EKSClusterActions describes the subset of methods from eksctl's cluster actions used here.
type EKSClusterActions interface {
	// Delete deletes an EKS cluster.
	//
	// Parameters:
	//   ctx - context for cancellation and deadlines.
	//   waitInterval - duration to wait between status checks.
	//   podEvictionWaitPeriod - duration to wait for pod eviction from nodegroups.
	//   wait - if true, waits for the cluster deletion to complete before returning.
	//   force - if true, forces deletion even if there are issues (e.g., stuck resources).
	//   disableNodegroupEviction - if true, skips eviction of pods from nodegroups before deletion.
	//   parallel - number of parallel operations to use during deletion.
	Delete(
		ctx context.Context,
		waitInterval, podEvictionWaitPeriod time.Duration,
		wait, force, disableNodegroupEviction bool,
		parallel int,
	) error
}

// EKSClusterActionsFactory describes the factory for creating cluster action instances.
type EKSClusterActionsFactory interface {
	// NewClusterActions creates cluster actions instance
	NewClusterActions(
		ctx context.Context,
		cfg *v1alpha5.ClusterConfig,
		ctl *eks.ClusterProvider,
	) (EKSClusterActions, error)
}

// EKSProviderConstructor describes the constructor for creating ClusterProvider instances.
type EKSProviderConstructor interface {
	// NewClusterProvider creates a new EKS ClusterProvider
	NewClusterProvider(
		ctx context.Context,
		spec *v1alpha5.ProviderConfig,
		clusterSpec *v1alpha5.ClusterConfig,
	) (*eks.ClusterProvider, error)
}

// EKSClusterLister describes the interface for listing clusters.
type EKSClusterLister interface {
	// GetClusters lists all EKS clusters
	GetClusters(
		ctx context.Context,
		provider *eks.ClusterProvider,
		listAllRegions bool,
		chunkSize int,
	) ([]cluster.Description, error)
}

// EKSClusterCreator describes the interface for creating clusters.
type EKSClusterCreator interface {
	// CreateCluster creates a new EKS cluster
	CreateCluster(
		ctx context.Context,
		cfg *v1alpha5.ClusterConfig,
		ctl *eks.ClusterProvider,
	) error
}

// EKSNodeGroupManager describes the interface for managing node groups.
type EKSNodeGroupManager interface {
	// Scale scales a node group to the specified size
	Scale(ctx context.Context, ng *v1alpha5.NodeGroupBase, wait bool) error
}

// EKSNodeGroupManagerFactory describes the factory for creating node group manager instances.
type EKSNodeGroupManagerFactory interface {
	// NewNodeGroupManager creates a node group manager instance
	NewNodeGroupManager(
		cfg *v1alpha5.ClusterConfig,
		ctl *eks.ClusterProvider,
		clientSet kubernetes.Interface,
		instanceSelector eks.InstanceSelector,
	) EKSNodeGroupManager
}