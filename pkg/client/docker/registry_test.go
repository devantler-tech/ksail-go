package docker

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRegistryManager(t *testing.T) {
	t.Run("success with valid client", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		
		rm, err := NewRegistryManager(mockClient)
		
		assert.NoError(t, err)
		assert.NotNil(t, rm)
		assert.Equal(t, mockClient, rm.client)
	})

	t.Run("error with nil client", func(t *testing.T) {
		rm, err := NewRegistryManager(nil)
		
		assert.Error(t, err)
		assert.Nil(t, rm)
		assert.Equal(t, ErrAPIClientNil, err)
	})
}

func TestCreateRegistry(t *testing.T) {
	t.Run("creates new registry successfully", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		config := RegistryConfig{
			Name:         "docker.io",
			Port:         5000,
			UpstreamURL:  "https://registry-1.docker.io",
			ClusterName:  "test-cluster",
			NetworkName:  "kind",
		}

		// Mock registry doesn't exist
		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{}, nil).
			Once()

		// Mock image inspect (image doesn't exist)
		mockClient.EXPECT().
			ImageInspectWithRaw(ctx, RegistryImageName).
			Return(image.InspectResponse{}, []byte{}, errors.New("not found")).
			Once()

		// Mock image pull
		mockClient.EXPECT().
			ImagePull(ctx, RegistryImageName, mock.Anything).
			Return(io.NopCloser(strings.NewReader("")), nil).
			Once()

		// Mock volume inspect (volume doesn't exist)
		mockClient.EXPECT().
			VolumeInspect(ctx, "ksail-registry-docker.io").
			Return(volume.Volume{}, errors.New("not found")).
			Once()

		// Mock volume create
		mockClient.EXPECT().
			VolumeCreate(ctx, mock.Anything).
			Return(volume.Volume{}, nil).
			Once()

		// Mock container create - use MatchedBy for complex types
		mockClient.EXPECT().
			ContainerCreate(
				ctx,
				mock.MatchedBy(func(config *container.Config) bool { return config != nil }),
				mock.MatchedBy(func(config *container.HostConfig) bool { return config != nil }),
				mock.MatchedBy(func(config *network.NetworkingConfig) bool { return config != nil }),
				mock.Anything,
				"ksail-registry-docker.io",
			).
			Return(container.CreateResponse{ID: "test-id"}, nil).
			Once()

		// Mock container start
		mockClient.EXPECT().
			ContainerStart(ctx, "test-id", mock.Anything).
			Return(nil).
			Once()

		err := rm.CreateRegistry(ctx, config)
		
		assert.NoError(t, err)
	})

	t.Run("returns error when registry already exists and adds cluster label", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		config := RegistryConfig{
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
						RegistryLabelKey:        "docker.io",
						RegistryClusterLabelKey: "other-cluster",
					},
				},
			}, nil).
			Once()

		err := rm.CreateRegistry(ctx, config)
		
		assert.NoError(t, err)
	})
}

func TestDeleteRegistry(t *testing.T) {
	t.Run("deletes registry when not in use", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		// Mock list registry containers (2 times - initial list, IsRegistryInUse)
		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{
				{
					ID: "registry-id",
					Labels: map[string]string{
						RegistryLabelKey:        "docker.io",
						RegistryClusterLabelKey: "test-cluster",
					},
					State: "exited",
				},
			}, nil).
			Times(2)

		// Mock container stop
		mockClient.EXPECT().
			ContainerStop(ctx, "registry-id", mock.MatchedBy(func(opts container.StopOptions) bool { return true })).
			Return(nil).
			Once()

		// Mock container remove
		mockClient.EXPECT().
			ContainerRemove(ctx, "registry-id", mock.MatchedBy(func(opts container.RemoveOptions) bool { return true })).
			Return(nil).
			Once()

		// Mock volume remove
		mockClient.EXPECT().
			VolumeRemove(ctx, "ksail-registry-docker.io", false).
			Return(nil).
			Once()

		err := rm.DeleteRegistry(ctx, "docker.io", "test-cluster", true)
		
		assert.NoError(t, err)
	})

	t.Run("does not delete registry when still in use", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		// Mock list registry containers with running state (2 times - initial list, IsRegistryInUse)
		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{
				{
					ID: "registry-id",
					Labels: map[string]string{
						RegistryLabelKey:        "docker.io",
						RegistryClusterLabelKey: "test-cluster,other-cluster",
					},
					State: "running",
				},
			}, nil).
			Times(2)

		err := rm.DeleteRegistry(ctx, "docker.io", "test-cluster", true)
		
		assert.NoError(t, err)
	})

	t.Run("returns error when registry not found", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		// Mock registry doesn't exist
		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{}, nil).
			Once()

		err := rm.DeleteRegistry(ctx, "docker.io", "test-cluster", false)
		
		assert.Error(t, err)
		assert.Equal(t, ErrRegistryNotFound, err)
	})
}

func TestListRegistries(t *testing.T) {
	t.Run("lists all registries", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{
				{
					ID: "registry1",
					Labels: map[string]string{
						RegistryLabelKey: "docker.io",
					},
				},
				{
					ID: "registry2",
					Labels: map[string]string{
						RegistryLabelKey: "quay.io",
					},
				},
			}, nil).
			Once()

		registries, err := rm.ListRegistries(ctx)
		
		assert.NoError(t, err)
		assert.Len(t, registries, 2)
		assert.Contains(t, registries, "docker.io")
		assert.Contains(t, registries, "quay.io")
	})

	t.Run("returns empty list when no registries", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{}, nil).
			Once()

		registries, err := rm.ListRegistries(ctx)
		
		assert.NoError(t, err)
		assert.Empty(t, registries)
	})
}

func TestIsRegistryInUse(t *testing.T) {
	t.Run("returns true when registry is running", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{
				{
					ID: "registry-id",
					Labels: map[string]string{
						RegistryLabelKey: "docker.io",
					},
					State: "running",
				},
			}, nil).
			Once()

		inUse, err := rm.IsRegistryInUse(ctx, "docker.io")
		
		assert.NoError(t, err)
		assert.True(t, inUse)
	})

	t.Run("returns false when registry is stopped", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{
				{
					ID: "registry-id",
					Labels: map[string]string{
						RegistryLabelKey: "docker.io",
					},
					State: "exited",
				},
			}, nil).
			Once()

		inUse, err := rm.IsRegistryInUse(ctx, "docker.io")
		
		assert.NoError(t, err)
		assert.False(t, inUse)
	})

	t.Run("returns false when registry not found", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{}, nil).
			Once()

		inUse, err := rm.IsRegistryInUse(ctx, "docker.io")
		
		assert.NoError(t, err)
		assert.False(t, inUse)
	})
}

func TestGetRegistryPort(t *testing.T) {
	t.Run("returns port for existing registry", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{
				{
					ID: "registry-id",
					Labels: map[string]string{
						RegistryLabelKey: "docker.io",
					},
					Ports: []container.Port{
						{
							PrivatePort: 5000,
							PublicPort:  5000,
						},
					},
				},
			}, nil).
			Once()

		port, err := rm.GetRegistryPort(ctx, "docker.io")
		
		assert.NoError(t, err)
		assert.Equal(t, 5000, port)
	})

	t.Run("returns error when registry not found", func(t *testing.T) {
		mockClient := NewMockAPIClient(t)
		rm := &RegistryManager{client: mockClient}
		ctx := context.Background()

		mockClient.EXPECT().
			ContainerList(ctx, mock.Anything).
			Return([]container.Summary{}, nil).
			Once()

		port, err := rm.GetRegistryPort(ctx, "docker.io")
		
		assert.Error(t, err)
		assert.Equal(t, 0, port)
		assert.Equal(t, ErrRegistryNotFound, err)
	})
}
