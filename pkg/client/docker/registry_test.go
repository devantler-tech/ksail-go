package docker_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	docker "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errNotFound = errors.New("not found")

// setupTestRegistryManager creates a test setup with mock client, manager, and context.
func setupTestRegistryManager(
	t *testing.T,
) (*docker.MockAPIClient, *docker.RegistryManager, context.Context) {
	t.Helper()

	mockClient := docker.NewMockAPIClient(t)
	manager, err := docker.NewRegistryManager(mockClient)
	require.NoError(t, err)

	ctx := context.Background()

	return mockClient, manager, ctx
}

// mockRegistryNotExists sets up mocks for when a registry doesn't exist.
func mockRegistryNotExists(ctx context.Context, mockClient *docker.MockAPIClient) {
	mockClient.EXPECT().
		ContainerList(ctx, mock.Anything).
		Return([]container.Summary{}, nil).
		Once()
}

// mockImagePullSequence sets up the complete image pull mock sequence.
func mockImagePullSequence(ctx context.Context, mockClient *docker.MockAPIClient) {
	// Mock image inspect (image doesn't exist)
	mockClient.EXPECT().
		ImageInspect(ctx, docker.RegistryImageName).
		Return(image.InspectResponse{}, errNotFound).
		Once()

	// Mock image pull
	mockClient.EXPECT().
		ImagePull(ctx, docker.RegistryImageName, mock.Anything).
		Return(io.NopCloser(strings.NewReader("")), nil).
		Once()
}

// mockVolumeCreateSequence sets up the complete volume creation mock sequence.
func mockVolumeCreateSequence(
	ctx context.Context,
	mockClient *docker.MockAPIClient,
	volumeName string,
) {
	// Mock volume inspect (volume doesn't exist)
	mockClient.EXPECT().
		VolumeInspect(ctx, volumeName).
		Return(volume.Volume{}, errNotFound).
		Once()

	// Mock volume create
	mockClient.EXPECT().
		VolumeCreate(ctx, mock.Anything).
		Return(volume.Volume{}, nil).
		Once()
}

// mockContainerCreateStart sets up the container creation and start mock sequence.
func mockContainerCreateStart(
	ctx context.Context, mockClient *docker.MockAPIClient, containerName, containerID string,
) {
	// Mock container create
	mockClient.EXPECT().
		ContainerCreate(
			ctx,
			mock.MatchedBy(func(config *container.Config) bool { return config != nil }),
			mock.MatchedBy(func(config *container.HostConfig) bool { return config != nil }),
			mock.MatchedBy(func(config *network.NetworkingConfig) bool { return config != nil }),
			mock.Anything,
			containerName,
		).
		Return(container.CreateResponse{ID: containerID}, nil).
		Once()

	// Mock container start
	mockClient.EXPECT().
		ContainerStart(ctx, containerID, mock.Anything).
		Return(nil).
		Once()
}

// mockRegistryContainer returns a mock container summary for a registry.
func mockRegistryContainer(registryID, registryName, clusterName, state string) container.Summary {
	return container.Summary{
		ID: registryID,
		Labels: map[string]string{
			docker.RegistryLabelKey:        registryName,
			docker.RegistryClusterLabelKey: clusterName,
		},
		State: state,
	}
}

// mockContainerStopRemove sets up the container stop and remove mock sequence.
func mockContainerStopRemove(
	ctx context.Context,
	mockClient *docker.MockAPIClient,
	containerID string,
) {
	mockClient.EXPECT().
		ContainerStop(ctx, containerID, mock.MatchedBy(func(_ container.StopOptions) bool { return true })).
		Return(nil).
		Once()

	mockClient.EXPECT().
		ContainerRemove(ctx, containerID, mock.MatchedBy(func(_ container.RemoveOptions) bool { return true })).
		Return(nil).
		Once()
}

// mockContainerListOnce sets up a single ContainerList mock with specified registry.
func mockContainerListOnce(
	ctx context.Context, mockClient *docker.MockAPIClient, registries []container.Summary,
) {
	mockClient.EXPECT().
		ContainerList(ctx, mock.Anything).
		Return(registries, nil).
		Once()
}

func TestNewRegistryManager(t *testing.T) {
	t.Parallel()

	t.Run("success with valid client", func(t *testing.T) {
		t.Parallel()

		mockClient := docker.NewMockAPIClient(t)

		manager, err := docker.NewRegistryManager(mockClient)

		require.NoError(t, err)
		require.NotNil(t, manager)
	})

	t.Run("error with nil client", func(t *testing.T) {
		t.Parallel()

		manager, err := docker.NewRegistryManager(nil)

		require.Error(t, err)
		assert.Nil(t, manager)
		assert.Equal(t, docker.ErrAPIClientNil, err)
	})
}

func TestCreateRegistry(t *testing.T) {
	t.Parallel()

	t.Run("creates new registry successfully", func(t *testing.T) {
		t.Parallel()

		mockClient, manager, ctx := setupTestRegistryManager(t)

		config := docker.RegistryConfig{
			Name:        "docker.io",
			Port:        5000,
			UpstreamURL: "https://registry-1.docker.io",
			ClusterName: "test-cluster",
			NetworkName: "kind",
		}

		mockRegistryNotExists(ctx, mockClient)
		mockImagePullSequence(ctx, mockClient)
		mockVolumeCreateSequence(ctx, mockClient, "ksail-registry-docker.io")
		mockContainerCreateStart(ctx, mockClient, "ksail-registry-docker.io", "test-id")

		err := manager.CreateRegistry(ctx, config)

		require.NoError(t, err)
	})

	t.Run("returns error when registry already exists and adds cluster label", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		config := docker.RegistryConfig{
			Name:        "docker.io",
			ClusterName: "test-cluster",
		}

		// Mock registry exists (called once in registryExists)
		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{
				{
					ID: "existing-id",
					Labels: map[string]string{
						docker.RegistryLabelKey:        "docker.io",
						docker.RegistryClusterLabelKey: "other-cluster",
					},
				},
			}, nil).
			Once()

		err := manager.CreateRegistry(ctx, config)

		require.NoError(t, err)
	})
}

func TestDeleteRegistry(t *testing.T) {
	t.Parallel()

	t.Run("deletes registry when not in use", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		registry := mockRegistryContainer("registry-id", "docker.io", "test-cluster", "exited")

		// Mock list registry containers (2 times - initial list, IsRegistryInUse)
		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{registry}, nil).
			Times(2)

		mockContainerStopRemove(ctx, mockClient, "registry-id")

		// Mock volume remove
		mockClient.EXPECT().
			VolumeRemove(ctx, "ksail-registry-docker.io", false).
			Return(nil).
			Once()

		err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", true)

		require.NoError(t, err)
	})

	t.Run("does not delete registry when still in use", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		registry := container.Summary{
			ID: "registry-id",
			Labels: map[string]string{
				docker.RegistryLabelKey:        "docker.io",
				docker.RegistryClusterLabelKey: "test-cluster,other-cluster",
			},
			State: "running",
		}

		// Mock list registry containers with running state (2 times - initial list, IsRegistryInUse)
		mockClient.EXPECT().ContainerList(ctx, mock.Anything).Return(
			[]container.Summary{registry}, nil,
		).Times(2)

		err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", true)

		require.NoError(t, err)
	})

	t.Run("returns error when registry not found", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		mockRegistryNotExists(ctx, mockClient)

		err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false)

		require.Error(t, err)
		assert.Equal(t, docker.ErrRegistryNotFound, err)
	})
}

func TestListRegistries(t *testing.T) {
	t.Parallel()

	t.Run("lists all registries", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		mockContainerListOnce(ctx, mockClient, []container.Summary{
			{
				ID: "registry1",
				Labels: map[string]string{
					docker.RegistryLabelKey: "docker.io",
				},
			},
			{
				ID: "registry2",
				Labels: map[string]string{
					docker.RegistryLabelKey: "quay.io",
				},
			},
		})

		registries, err := manager.ListRegistries(ctx)

		require.NoError(t, err)
		assert.Len(t, registries, 2)
		assert.Contains(t, registries, "docker.io")
		assert.Contains(t, registries, "quay.io")
	})

	t.Run("returns empty list when no registries", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		mockRegistryNotExists(ctx, mockClient)

		registries, err := manager.ListRegistries(ctx)

		require.NoError(t, err)
		assert.Empty(t, registries)
	})
}

func TestIsRegistryInUse(t *testing.T) {
	t.Parallel()

	t.Run("returns true when registry is running", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		registry := mockRegistryContainer("registry-id", "docker.io", "", "running")
		registry.Labels = map[string]string{
			docker.RegistryLabelKey: "docker.io",
		} // Simplified labels

		mockContainerListOnce(ctx, mockClient, []container.Summary{registry})

		inUse, err := manager.IsRegistryInUse(ctx, "docker.io")

		require.NoError(t, err)
		assert.True(t, inUse)
	})

	t.Run("returns false when registry is stopped", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		registry := mockRegistryContainer("registry-id", "docker.io", "", "exited")
		registry.Labels = map[string]string{docker.RegistryLabelKey: "docker.io"}

		mockContainerListOnce(ctx, mockClient, []container.Summary{registry})

		inUse, err := manager.IsRegistryInUse(ctx, "docker.io")

		require.NoError(t, err)
		assert.False(t, inUse)
	})

	t.Run("returns false when registry not found", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		mockRegistryNotExists(ctx, mockClient)

		inUse, err := manager.IsRegistryInUse(ctx, "docker.io")

		require.NoError(t, err)
		assert.False(t, inUse)
	})
}

func TestGetRegistryPort(t *testing.T) {
	t.Parallel()

	t.Run("returns port for existing registry", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		mockContainerListOnce(ctx, mockClient, []container.Summary{
			{
				ID: "registry-id",
				Labels: map[string]string{
					docker.RegistryLabelKey: "docker.io",
				},
				Ports: []container.Port{
					{
						PrivatePort: 5000,
						PublicPort:  5000,
					},
				},
			},
		})

		port, err := manager.GetRegistryPort(ctx, "docker.io")

		require.NoError(t, err)
		assert.Equal(t, 5000, port)
	})

	t.Run("returns error when registry not found", func(t *testing.T) {
		t.Parallel()
		mockClient, manager, ctx := setupTestRegistryManager(t)

		mockRegistryNotExists(ctx, mockClient)

		port, err := manager.GetRegistryPort(ctx, "docker.io")

		require.Error(t, err)
		assert.Equal(t, 0, port)
		assert.Equal(t, docker.ErrRegistryNotFound, err)
	})
}
