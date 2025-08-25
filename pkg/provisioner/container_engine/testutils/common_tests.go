// Package testutils provides common test utilities for container engine provisioners.
package testutils

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	containerengineprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

// ProvisionerFactory is a function type that creates a provisioner with a mock client.
type ProvisionerFactory func(*provisioner.MockAPIClient) containerengineprovisioner.ContainerEngineProvisioner

// TestCheckReadySuccess runs a common test pattern for CheckReady success scenarios.
func TestCheckReadySuccess(t *testing.T, factory ProvisionerFactory) {
	t.Helper()

	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)
	provisioner := factory(mockClient)

	expectedPing := types.Ping{
		APIVersion:     "1.41",
		OSType:         "",
		Experimental:   false,
		SwarmStatus:    nil,
		BuilderVersion: "",
	}
	mockClient.On("Ping", mock.MatchedBy(func(_ context.Context) bool {
		return true
	})).Return(expectedPing, nil)

	// Act
	ready, err := provisioner.CheckReady()

	// Assert
	require.NoError(t, err)
	assert.True(t, ready)
	mockClient.AssertExpectations(t)
}

// TestCheckReadyError runs a common test pattern for CheckReady error scenarios.
func TestCheckReadyError(t *testing.T, factory ProvisionerFactory, expectedErrorSubstring string) {
	t.Helper()

	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)
	provisioner := factory(mockClient)

	mockClient.On("Ping", mock.MatchedBy(func(_ context.Context) bool {
		return true
	})).Return(types.Ping{
		APIVersion:     "",
		OSType:         "",
		Experimental:   false,
		BuilderVersion: "",
		SwarmStatus:    nil,
	}, errBoom)

	// Act
	ready, err := provisioner.CheckReady()

	// Assert
	require.Error(t, err)
	assert.False(t, ready)
	assert.Contains(t, err.Error(), expectedErrorSubstring)
	assert.Contains(t, err.Error(), "boom")
	mockClient.AssertExpectations(t)
}


