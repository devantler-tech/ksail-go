package stubs

import (
	"context"

	"github.com/weaveworks/eksctl/pkg/actions/cluster"
	"github.com/weaveworks/eksctl/pkg/eks"
)

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
func (e *EKSClusterListerStub) GetClusters(
	ctx context.Context,
	provider *eks.ClusterProvider,
	listAllRegions bool,
	chunkSize int,
) ([]cluster.Description, error) {
	e.GetClustersCalls++
	return e.GetClustersResult, e.GetClustersError
}
