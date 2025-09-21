package eksprovisioner

import (
	"context"
	"time"

	"github.com/weaveworks/eksctl/pkg/actions/cluster"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/eks"
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
