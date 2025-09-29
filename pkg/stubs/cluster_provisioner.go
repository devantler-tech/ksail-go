package stubs

import (
	"context"
	"errors"
)

// ClusterProvisionerStub is a stub implementation of clusterprovisioner.ClusterProvisioner interface.
// It provides configurable behavior for testing without external dependencies.
type ClusterProvisionerStub struct {
	CreateError  error
	DeleteError  error
	StartError   error
	StopError    error
	ListResult   []string
	ListError    error
	ExistsResult bool
	ExistsError  error
	
	// Track calls for verification
	CreateCalls []string
	DeleteCalls []string
	StartCalls  []string
	StopCalls   []string
	ListCalls   int
	ExistsCalls []string
}

// NewClusterProvisionerStub creates a new ClusterProvisionerStub with default behavior.
func NewClusterProvisionerStub() *ClusterProvisionerStub {
	return &ClusterProvisionerStub{
		ListResult:   []string{"test-cluster"},
		ExistsResult: true,
	}
}

// Create simulates cluster creation.
func (c *ClusterProvisionerStub) Create(ctx context.Context, name string) error {
	c.CreateCalls = append(c.CreateCalls, name)
	return c.CreateError
}

// Delete simulates cluster deletion.
func (c *ClusterProvisionerStub) Delete(ctx context.Context, name string) error {
	c.DeleteCalls = append(c.DeleteCalls, name)
	return c.DeleteError
}

// Start simulates cluster start.
func (c *ClusterProvisionerStub) Start(ctx context.Context, name string) error {
	c.StartCalls = append(c.StartCalls, name)
	return c.StartError
}

// Stop simulates cluster stop.
func (c *ClusterProvisionerStub) Stop(ctx context.Context, name string) error {
	c.StopCalls = append(c.StopCalls, name)
	return c.StopError
}

// List simulates cluster listing.
func (c *ClusterProvisionerStub) List(ctx context.Context) ([]string, error) {
	c.ListCalls++
	if c.ListError != nil {
		return nil, c.ListError
	}
	return c.ListResult, nil
}

// Exists simulates cluster existence check.
func (c *ClusterProvisionerStub) Exists(ctx context.Context, name string) (bool, error) {
	c.ExistsCalls = append(c.ExistsCalls, name)
	return c.ExistsResult, c.ExistsError
}

// WithCreateError configures the stub to return an error on Create.
func (c *ClusterProvisionerStub) WithCreateError(message string) *ClusterProvisionerStub {
	c.CreateError = errors.New(message)
	return c
}

// WithDeleteError configures the stub to return an error on Delete.
func (c *ClusterProvisionerStub) WithDeleteError(message string) *ClusterProvisionerStub {
	c.DeleteError = errors.New(message)
	return c
}

// WithStartError configures the stub to return an error on Start.
func (c *ClusterProvisionerStub) WithStartError(message string) *ClusterProvisionerStub {
	c.StartError = errors.New(message)
	return c
}

// WithStopError configures the stub to return an error on Stop.
func (c *ClusterProvisionerStub) WithStopError(message string) *ClusterProvisionerStub {
	c.StopError = errors.New(message)
	return c
}

// WithListResult configures the stub to return specific clusters on List.
func (c *ClusterProvisionerStub) WithListResult(clusters []string) *ClusterProvisionerStub {
	c.ListResult = clusters
	c.ListError = nil
	return c
}

// WithListError configures the stub to return an error on List.
func (c *ClusterProvisionerStub) WithListError(message string) *ClusterProvisionerStub {
	c.ListError = errors.New(message)
	return c
}

// WithExistsResult configures the stub to return a specific result on Exists.
func (c *ClusterProvisionerStub) WithExistsResult(exists bool) *ClusterProvisionerStub {
	c.ExistsResult = exists
	c.ExistsError = nil
	return c
}

// WithExistsError configures the stub to return an error on Exists.
func (c *ClusterProvisionerStub) WithExistsError(message string) *ClusterProvisionerStub {
	c.ExistsError = errors.New(message)
	return c
}

// GetCreateCallsCount returns the number of Create calls.
func (c *ClusterProvisionerStub) GetCreateCallsCount() int {
	return len(c.CreateCalls)
}

// GetLastCreateCall returns the last name passed to Create.
func (c *ClusterProvisionerStub) GetLastCreateCall() string {
	if len(c.CreateCalls) == 0 {
		return ""
	}
	return c.CreateCalls[len(c.CreateCalls)-1]
}

// Reset clears all call tracking.
func (c *ClusterProvisionerStub) Reset() {
	c.CreateCalls = nil
	c.DeleteCalls = nil
	c.StartCalls = nil
	c.StopCalls = nil
	c.ListCalls = 0
	c.ExistsCalls = nil
}