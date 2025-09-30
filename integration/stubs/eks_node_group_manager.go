package stubs

import (
	"context"

	eksapi "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

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
func (e *EKSNodeGroupManagerStub) Scale(
	ctx context.Context,
	ng *eksapi.NodeGroupBase,
	wait bool,
) error {
	e.ScaleCalls++
	return e.ScaleError
}
