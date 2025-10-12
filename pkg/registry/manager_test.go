package registry_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	docker "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/devantler-tech/ksail-go/pkg/registry"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupManagerTest creates a test context, config, and mock client.
func setupManagerTest(t *testing.T) (
	context.Context,
	registry.Config,
	*docker.MockAPIClient,
) {
	t.Helper()

	mockClient := docker.NewMockAPIClient(t)
	ctx := context.Background()
	cfg := registry.Config{
		Name:     "test-registry",
		HostPort: "5000",
	}

	return ctx, cfg, mockClient
}

// mockContainerDoesNotExist sets up mock expectations for a container that doesn't exist.
func mockContainerDoesNotExist(ctx context.Context, mockClient *docker.MockAPIClient) {
	mockClient.On("ContainerList", ctx, mock.Anything).Return([]container.Summary{}, nil)
}

// mockSuccessfulImagePull sets up mock expectations for successful image pull.
func mockSuccessfulImagePull(
	ctx context.Context,
	mockClient *docker.MockAPIClient,
	image string,
) {
	mockClient.On("ImagePull", ctx, image, mock.Anything).
		Return(io.NopCloser(strings.NewReader("")), nil)
}

// mockSuccessfulContainerCreate sets up mock expectations for successful container creation.
func mockSuccessfulContainerCreate(
	ctx context.Context,
	mockClient *docker.MockAPIClient,
	containerName string,
) {
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, containerName).
		Return(container.CreateResponse{ID: "test-id"}, nil)
	mockClient.On("ContainerStart", ctx, "test-id", mock.Anything).Return(nil)
}

// TestCreateRegistry_Success tests successful registry creation.
func TestCreateRegistrySuccess(t *testing.T) {
	t.Parallel()

	ctx, cfg, mockClient := setupManagerTest(t)
	cfg.Image = registry.DefaultRegistryImage

	mockContainerDoesNotExist(ctx, mockClient)
	mockSuccessfulImagePull(ctx, mockClient, cfg.Image)
	mockSuccessfulContainerCreate(ctx, mockClient, cfg.Name)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.NoError(t, err)
}

// TestCreateRegistry_AlreadyExists tests registry creation when container already exists.
func TestCreateRegistryAlreadyExists(t *testing.T) {
	t.Parallel()

	ctx, cfg, mockClient := setupManagerTest(t)

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

	ctx, cfg, mockClient := setupManagerTest(t)
	errTestListFailed := errors.New("list error") //nolint:err113 // Test error

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

	ctx, cfg, mockClient := setupManagerTest(t)
	// Image not specified - should use default

	mockContainerDoesNotExist(ctx, mockClient)
	mockSuccessfulImagePull(ctx, mockClient, registry.DefaultRegistryImage)
	mockSuccessfulContainerCreate(ctx, mockClient, cfg.Name)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.NoError(t, err)
}

// TestCreateRegistry_PullError tests error handling when image pull fails.
func TestCreateRegistryPullError(t *testing.T) {
	t.Parallel()

	ctx, cfg, mockClient := setupManagerTest(t)
	errTestPullFailed := errors.New("pull error") //nolint:err113 // Test error

	mockContainerDoesNotExist(ctx, mockClient)
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

	ctx, cfg, mockClient := setupManagerTest(t)
	errTestCreateFailed := errors.New("create error") //nolint:err113 // Test error

	mockContainerDoesNotExist(ctx, mockClient)
	mockSuccessfulImagePull(ctx, mockClient, registry.DefaultRegistryImage)
	mockClient.On("ContainerCreate", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, cfg.Name).
		Return(container.CreateResponse{}, errTestCreateFailed)

	manager := registry.NewManager(mockClient)
	err := manager.CreateRegistry(ctx, cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create registry container")
}
