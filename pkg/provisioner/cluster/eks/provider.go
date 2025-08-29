// Package eksprovisioner provides implementations of the Provisioner interface
// for provisioning EKS clusters in AWS.
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
	// Delete deletes an EKS cluster
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