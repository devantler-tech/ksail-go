package stubs

import (
	"context"
	"time"
)

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