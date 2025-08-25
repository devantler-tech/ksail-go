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

// TestCheckReadySuccess tests the CheckReady method for successful scenarios.
func TestCheckReadySuccess(
	t *testing.T,
	provisioner containerengineprovisioner.ContainerEngineProvisioner,
	mockClient *provisioner.MockAPIClient,
) {
	t.Helper()

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

// TestCheckReadyError tests the CheckReady method for error scenarios.
func TestCheckReadyError(
	t *testing.T,
	provisioner containerengineprovisioner.ContainerEngineProvisioner,
	mockClient *provisioner.MockAPIClient,
	expectedErrorSubstring string,
) {
	t.Helper()

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
