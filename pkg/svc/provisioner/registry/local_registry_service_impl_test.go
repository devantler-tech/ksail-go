package registry

import (
	"context"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRegistryBackend struct {
	mock.Mock
}

func (m *mockRegistryBackend) CreateRegistry(
	ctx context.Context,
	cfg dockerclient.RegistryConfig,
) error {
	args := m.Called(ctx, cfg)

	return args.Error(0)
}

func (m *mockRegistryBackend) DeleteRegistry(
	ctx context.Context,
	name, clusterName string,
	deleteVolume bool,
	networkName string,
) error {
	args := m.Called(ctx, name, clusterName, deleteVolume, networkName)

	return args.Error(0)
}

func (m *mockRegistryBackend) ListRegistries(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)

	if registries, ok := args.Get(0).([]string); ok {
		return registries, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *mockRegistryBackend) GetRegistryPort(ctx context.Context, name string) (int, error) {
	args := m.Called(ctx, name)

	return args.Int(0), args.Error(1)
}

func TestNewServiceRequiresDockerClient(t *testing.T) {
	_, err := NewService(Config{})
	require.Error(t, err)
}

func TestCreateEnsuresRegistryMetadata(t *testing.T) {
	t.Parallel()

	docker := dockerclient.NewMockAPIClient(t)
	backend := &mockRegistryBackend{}

	backend.
		On("ListRegistries", mock.Anything).
		Return([]string{}, nil).
		Once()

	backend.
		On(
			"CreateRegistry",
			mock.Anything,
			mock.MatchedBy(func(cfg dockerclient.RegistryConfig) bool {
				return cfg.Name == "local-registry" && cfg.Port == 5000 &&
					cfg.VolumeName == "local-registry"
			}),
		).
		Return(nil).
		Once()

	docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{runningSummary()}, nil).
		Once()

	svc, err := NewService(Config{
		DockerClient:    docker,
		RegistryManager: backend,
	})
	require.NoError(t, err)

	registry, err := svc.Create(
		context.Background(),
		CreateOptions{Name: "local-registry", Port: 5000},
	)
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1:5000", registry.Endpoint)
	assert.Equal(t, int32(5000), registry.Port)
	assert.Equal(t, "local-registry-data", registry.VolumeName)
	assert.Equal(t, v1alpha1.OCIRegistryStatusRunning, registry.Status)
}

func TestStartStartsAndConnectsRegistry(t *testing.T) {
	t.Parallel()

	docker := dockerclient.NewMockAPIClient(t)
	backend := &mockRegistryBackend{}

	svc, err := NewService(Config{DockerClient: docker, RegistryManager: backend})
	require.NoError(t, err)

	docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{exitedSummary()}, nil).
		Once()

	docker.
		On("ContainerStart", mock.Anything, "registry-id", mock.Anything).
		Return(nil).
		Once()

	docker.
		On("ContainerInspect", mock.Anything, "registry-id").
		Return(container.InspectResponse{NetworkSettings: &types.NetworkSettings{Networks: map[string]*network.EndpointSettings{}}}, nil).
		Once()

	docker.
		On("NetworkConnect", mock.Anything, "kind", "registry-id", mock.Anything).
		Return(nil).
		Once()

	docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{runningSummary()}, nil).
		Once()

	registry, err := svc.Start(
		context.Background(),
		StartOptions{Name: "local-registry", NetworkName: "kind"},
	)
	require.NoError(t, err)
	assert.Equal(t, v1alpha1.OCIRegistryStatusRunning, registry.Status)
}

func TestStopDeletesResourcesWhenRequested(t *testing.T) {
	t.Parallel()

	docker := dockerclient.NewMockAPIClient(t)
	backend := &mockRegistryBackend{}

	svc, err := NewService(Config{DockerClient: docker, RegistryManager: backend})
	require.NoError(t, err)

	backend.
		On("DeleteRegistry", mock.Anything, "local-registry", "dev", true, "kind").
		Return(nil).
		Once()

	require.NoError(t, svc.Stop(context.Background(), StopOptions{
		Name:         "local-registry",
		ClusterName:  "dev",
		NetworkName:  "kind",
		DeleteVolume: true,
	}))
}

func TestStopGracefullyStopsContainer(t *testing.T) {
	t.Parallel()

	docker := dockerclient.NewMockAPIClient(t)
	backend := &mockRegistryBackend{}

	svc, err := NewService(Config{DockerClient: docker, RegistryManager: backend})
	require.NoError(t, err)

	docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{runningSummary()}, nil).
		Once()

	docker.
		On("ContainerStop", mock.Anything, "registry-id", mock.Anything).
		Return(nil).
		Once()

	docker.
		On("NetworkDisconnect", mock.Anything, "kind", "registry-id", true).
		Return(nil).
		Once()

	require.NoError(
		t,
		svc.Stop(context.Background(), StopOptions{Name: "local-registry", NetworkName: "kind"}),
	)
}

func TestStatusReturnsNotProvisionedWhenMissing(t *testing.T) {
	t.Parallel()

	docker := dockerclient.NewMockAPIClient(t)
	backend := &mockRegistryBackend{}

	svc, err := NewService(Config{DockerClient: docker, RegistryManager: backend})
	require.NoError(t, err)

	docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{}, nil).
		Once()

	registry, err := svc.Status(context.Background(), StatusOptions{Name: "local-registry"})
	require.NoError(t, err)
	assert.Equal(t, v1alpha1.OCIRegistryStatusNotProvisioned, registry.Status)
}

func runningSummary() container.Summary {
	return container.Summary{
		ID:    "registry-id",
		State: "running",
		Ports: []types.Port{
			{PrivatePort: dockerclient.DefaultRegistryPort, PublicPort: 5000, IP: "127.0.0.1"},
		},
		Mounts: []types.MountPoint{{Type: "volume", Name: "local-registry-data"}},
		Labels: map[string]string{dockerclient.RegistryLabelKey: "local-registry"},
	}
}

func exitedSummary() container.Summary {
	summary := runningSummary()
	summary.State = "exited"

	return summary
}
