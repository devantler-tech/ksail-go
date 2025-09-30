package stubs

import (
	"context"

	eksapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"github.com/weaveworks/eksctl/pkg/eks"
)

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