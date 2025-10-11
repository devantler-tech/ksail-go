package registry_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/containerengine"
	"github.com/devantler-tech/ksail-go/pkg/registry"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	errTestListFailed   = errors.New("list error")
	errTestPullFailed   = errors.New("pull error")
	errTestCreateFailed = errors.New("create error")
)

// TestCreateRegistry_Success tests successful registry creation.
func TestCreateRegistrySuccess(t *testing.T) {
	t.Parallel()

	mockClient := containerengine.NewMockAPIClient(t)
	ctx := context.Background()

	cfg := registry.Config{
		Name:     "test-registry",
		HostPort: "5000",
		Image:    registry.DefaultRegistryImage,
	}

	// Mock container doesn't exist
	mockClient.On("ContainerList", ctx, mock.Anything).Return([]container.Summary{}, nil)

	// Mock successful image pull
	mockClient.On("ImagePull", ctx, cfg.Image, mock.Anything).
		Return(io.NopCloser(strings.NewReader("")), nil)

	// Mock successful container create
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, cfg.Name).
		Return(container.CreateResponse{ID: "test-id"}, nil)

	// Mock successful container start
	mockClient.On("ContainerStart", ctx, "test-id", mock.Anything).Return(nil)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.NoError(t, err)
}

// TestCreateRegistry_AlreadyExists tests registry creation when container already exists.
func TestCreateRegistryAlreadyExists(t *testing.T) {
	t.Parallel()

	mockClient := containerengine.NewMockAPIClient(t)
	ctx := context.Background()

	cfg := registry.Config{
		Name:     "test-registry",
		HostPort: "5000",
	}

	// Mock container exists
	existingContainers := []container.Summary{
		{Names: []string{"/test-registry"}},
	}
	mockClient.On("ContainerList", ctx, mock.Anything).Return(existingContainers, nil)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.NoError(t, err)
	// Should not attempt to create or pull
	mockClient.AssertNotCalled(t, "ImagePull")
	mockClient.AssertNotCalled(t, "ContainerCreate")
}

// TestCreateRegistry_ListError tests error handling when listing containers fails.
func TestCreateRegistryListError(t *testing.T) {
	t.Parallel()

	mockClient := containerengine.NewMockAPIClient(t)
	ctx := context.Background()

	cfg := registry.Config{
		Name:     "test-registry",
		HostPort: "5000",
	}

	// Mock list error
	mockClient.On("ContainerList", ctx, mock.Anything).
		Return([]container.Summary{}, errTestListFailed)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check if registry exists")
}

// TestCreateRegistry_DefaultImage tests that default image is used when not specified.
func TestCreateRegistryDefaultImage(t *testing.T) {
	t.Parallel()

	mockClient := containerengine.NewMockAPIClient(t)
	ctx := context.Background()

	cfg := registry.Config{
		Name:     "test-registry",
		HostPort: "5000",
		// Image not specified
	}

	// Mock container doesn't exist
	mockClient.On("ContainerList", ctx, mock.Anything).Return([]container.Summary{}, nil)

	// Mock successful image pull with default image
	mockClient.On("ImagePull", ctx, registry.DefaultRegistryImage, mock.Anything).
		Return(io.NopCloser(strings.NewReader("")), nil)

	// Mock successful container create
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, cfg.Name).
		Return(container.CreateResponse{ID: "test-id"}, nil)

	// Mock successful container start
	mockClient.On("ContainerStart", ctx, "test-id", mock.Anything).Return(nil)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.NoError(t, err)
}

// TestCreateRegistry_PullError tests error handling when image pull fails.
func TestCreateRegistryPullError(t *testing.T) {
	t.Parallel()

	mockClient := containerengine.NewMockAPIClient(t)
	ctx := context.Background()

	cfg := registry.Config{
		Name:     "test-registry",
		HostPort: "5000",
	}

	// Mock container doesn't exist
	mockClient.On("ContainerList", ctx, mock.Anything).Return([]container.Summary{}, nil)

	// Mock image pull error
	mockClient.On("ImagePull", ctx, registry.DefaultRegistryImage, mock.Anything).
		Return(nil, errTestPullFailed)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pull registry image")
}

// TestCreateRegistry_CreateError tests error handling when container creation fails.
func TestCreateRegistryCreateError(t *testing.T) {
	t.Parallel()

	mockClient := containerengine.NewMockAPIClient(t)
	ctx := context.Background()

	cfg := registry.Config{
		Name:     "test-registry",
		HostPort: "5000",
	}

	// Mock container doesn't exist
	mockClient.On("ContainerList", ctx, mock.Anything).Return([]container.Summary{}, nil)

	// Mock successful image pull
	mockClient.On("ImagePull", ctx, registry.DefaultRegistryImage, mock.Anything).
		Return(io.NopCloser(strings.NewReader("")), nil)

	// Mock container create error
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, cfg.Name).
		Return(container.CreateResponse{}, errTestCreateFailed)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create registry container")
}
