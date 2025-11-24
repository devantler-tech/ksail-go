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

type registryTestHarness struct {
	docker  *dockerclient.MockAPIClient
	backend *mockRegistryBackend
	svc     Service
}

func newRegistryTestHarness(t *testing.T) registryTestHarness {
	t.Helper()

	docker := dockerclient.NewMockAPIClient(t)
	backend := &mockRegistryBackend{}

	svc, err := NewService(Config{DockerClient: docker, RegistryManager: backend})
	require.NoError(t, err)

	return registryTestHarness{docker: docker, backend: backend, svc: svc}
}

func TestCreateOptionsWithDefaults(t *testing.T) {
	t.Parallel()

	opts := CreateOptions{
		Name: "dev-cluster-registry",
		Port: 5000,
	}

	defaulted := opts.WithDefaults()

	require.Equal(t, DefaultEndpointHost, defaulted.Host)
	require.Equal(t, opts.Name, defaulted.VolumeName)
}

func TestCreateOptionsValidate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		opts    CreateOptions
		wantErr error
	}{
		{
			name: "valid configuration",
			opts: CreateOptions{Name: "ksail", Port: 5000},
		},
		{
			name:    "missing name",
			opts:    CreateOptions{Port: 5000},
			wantErr: ErrNameRequired,
		},
		{
			name:    "invalid port",
			opts:    CreateOptions{Name: "ksail", Port: 70000},
			wantErr: ErrInvalidPort,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.opts.Validate()
			if tc.wantErr == nil {
				require.NoError(t, err)

				return
			}

			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestCreateOptionsEndpoint(t *testing.T) {
	t.Parallel()

	opts := CreateOptions{Name: "ksail", Port: 5000}
	endpoint := opts.Endpoint()

	require.Equal(t, "localhost:5000", endpoint)
}

func TestStartStopStatusOptionValidation(t *testing.T) {
	t.Parallel()

	require.NoError(t, (StartOptions{Name: "ksail"}).Validate())
	require.ErrorIs(t, (StartOptions{}).Validate(), ErrNameRequired)

	require.NoError(t, (StopOptions{Name: "ksail"}).Validate())
	require.ErrorIs(t, (StopOptions{}).Validate(), ErrNameRequired)

	require.NoError(t, (StatusOptions{Name: "ksail"}).Validate())
	require.ErrorIs(t, (StatusOptions{}).Validate(), ErrNameRequired)
}

func TestNewServiceRequiresDockerClient(t *testing.T) {
	_, err := NewService(Config{})
	require.Error(t, err)
}

func TestCreateEnsuresRegistryMetadata(t *testing.T) {
	t.Parallel()

	h := newRegistryTestHarness(t)

	h.backend.
		On("ListRegistries", mock.Anything).
		Return([]string{}, nil).
		Once()

	h.backend.
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

	h.docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{runningSummary()}, nil).
		Once()

	registry, err := h.svc.Create(
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

	h := newRegistryTestHarness(t)

	h.docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{exitedSummary()}, nil).
		Once()

	h.docker.
		On("ContainerStart", mock.Anything, "registry-id", mock.Anything).
		Return(nil).
		Once()

	h.docker.
		On("ContainerInspect", mock.Anything, "registry-id").
		Return(container.InspectResponse{NetworkSettings: &types.NetworkSettings{Networks: map[string]*network.EndpointSettings{}}}, nil).
		Once()

	h.docker.
		On("NetworkConnect", mock.Anything, "kind", "registry-id", mock.Anything).
		Return(nil).
		Once()

	h.docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{runningSummary()}, nil).
		Once()

	registry, err := h.svc.Start(
		context.Background(),
		StartOptions{Name: "local-registry", NetworkName: "kind"},
	)
	require.NoError(t, err)
	assert.Equal(t, v1alpha1.OCIRegistryStatusRunning, registry.Status)
}

func TestStopDeletesResourcesWhenRequested(t *testing.T) {
	t.Parallel()

	h := newRegistryTestHarness(t)

	h.backend.
		On("DeleteRegistry", mock.Anything, "local-registry", "dev", true, "kind").
		Return(nil).
		Once()

	require.NoError(t, h.svc.Stop(context.Background(), StopOptions{
		Name:         "local-registry",
		ClusterName:  "dev",
		NetworkName:  "kind",
		DeleteVolume: true,
	}))
}

func TestStopGracefullyStopsContainer(t *testing.T) {
	t.Parallel()

	h := newRegistryTestHarness(t)

	h.docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{runningSummary()}, nil).
		Once()

	h.docker.
		On("ContainerStop", mock.Anything, "registry-id", mock.Anything).
		Return(nil).
		Once()

	h.docker.
		On("NetworkDisconnect", mock.Anything, "kind", "registry-id", true).
		Return(nil).
		Once()

	require.NoError(
		t,
		h.svc.Stop(context.Background(), StopOptions{Name: "local-registry", NetworkName: "kind"}),
	)
}

func TestStatusReturnsNotProvisionedWhenMissing(t *testing.T) {
	t.Parallel()

	h := newRegistryTestHarness(t)

	h.docker.
		On("ContainerList", mock.Anything, mock.Anything).
		Return([]container.Summary{}, nil).
		Once()

	registry, err := h.svc.Status(context.Background(), StatusOptions{Name: "local-registry"})
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
