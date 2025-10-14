package docker_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/docker"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// assertRegistryConfig validates a registry configuration.
func assertRegistryConfig(
	t *testing.T,
	reg docker.RegistryConfig,
	expectedName, expectedPort string,
) {
	t.Helper()

	assert.Equal(t, expectedName, reg.Name)
	assert.Equal(t, expectedPort, reg.HostPort)
	assert.Equal(t, docker.DefaultRegistryImage, reg.Image)
}

// assertKindSingleRegistry is a helper that validates extraction of a single registry from Kind config.
func assertKindSingleRegistry(
	t *testing.T,
	patch string,
	expectedName, expectedPort string,
) {
	t.Helper()

	cfg := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{patch},
	}

	registries, err := docker.ExtractRegistriesFromKind(cfg)

	require.NoError(t, err)
	require.Len(t, registries, 1)
	assertRegistryConfig(t, registries[0], expectedName, expectedPort)
}

func TestExtractRegistriesFromK3d_NoRegistries(t *testing.T) {
	t.Parallel()

	cfg := &k3dv1alpha5.SimpleConfig{}

	registries, err := docker.ExtractRegistriesFromK3d(cfg)

	require.NoError(t, err)
	assert.Empty(t, registries)
}

func TestExtractRegistriesFromK3d_SingleRegistry(t *testing.T) {
	t.Parallel()

	cfg := &k3dv1alpha5.SimpleConfig{
		Registries: k3dv1alpha5.SimpleConfigRegistries{
			Use: []string{"k3d-registry:5000"},
		},
	}

	registries, err := docker.ExtractRegistriesFromK3d(cfg)

	require.NoError(t, err)
	require.Len(t, registries, 1)
	assert.Equal(t, "k3d-registry", registries[0].Name)
	assert.Equal(t, "5000", registries[0].HostPort)
	assert.Equal(t, docker.DefaultRegistryImage, registries[0].Image)
}

func TestExtractRegistriesFromK3d_MultipleRegistries(t *testing.T) {
	t.Parallel()

	cfg := &k3dv1alpha5.SimpleConfig{
		Registries: k3dv1alpha5.SimpleConfigRegistries{
			Use: []string{"registry1:5000", "registry2:5001"},
		},
	}

	registries, err := docker.ExtractRegistriesFromK3d(cfg)

	require.NoError(t, err)
	require.Len(t, registries, 2)
	assertRegistryConfig(t, registries[0], "registry1", "5000")
	assertRegistryConfig(t, registries[1], "registry2", "5001")
}

func TestExtractRegistriesFromK3d_InvalidFormat(t *testing.T) {
	t.Parallel()

	cfg := &k3dv1alpha5.SimpleConfig{
		Registries: k3dv1alpha5.SimpleConfigRegistries{
			Use: []string{"invalid-format"},
		},
	}

	registries, err := docker.ExtractRegistriesFromK3d(cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid registry reference format")
	assert.Nil(t, registries)
}

func TestExtractRegistriesFromK3d_NilConfig(t *testing.T) {
	t.Parallel()

	registries, err := docker.ExtractRegistriesFromK3d(nil)

	require.NoError(t, err)
	assert.Empty(t, registries)
}

func TestExtractRegistriesFromKind_NoPatches(t *testing.T) {
	t.Parallel()

	cfg := &v1alpha4.Cluster{}

	registries, err := docker.ExtractRegistriesFromKind(cfg)

	require.NoError(t, err)
	assert.Empty(t, registries)
}

func TestExtractRegistriesFromKind_SingleRegistry(t *testing.T) {
	t.Parallel()

	patch := `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
  endpoint = ["http://k3d-registry:5000"]`

	assertKindSingleRegistry(t, patch, "k3d-registry", "5000")
}

func TestExtractRegistriesFromKind_MultiplePatches(t *testing.T) {
	t.Parallel()

	patch1 := `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
  endpoint = ["http://registry1:5000"]`

	patch2 := `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5001"]
  endpoint = ["http://registry2:5001"]`

	cfg := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{patch1, patch2},
	}

	registries, err := docker.ExtractRegistriesFromKind(cfg)

	require.NoError(t, err)
	require.Len(t, registries, 2)
	assert.Equal(t, "registry1", registries[0].Name)
	assert.Equal(t, "5000", registries[0].HostPort)
	assert.Equal(t, "registry2", registries[1].Name)
	assert.Equal(t, "5001", registries[1].HostPort)
}

func TestExtractRegistriesFromKind_DuplicateRegistries(t *testing.T) {
	t.Parallel()

	patch := `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
  endpoint = ["http://registry:5000"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry.local:5000"]
  endpoint = ["http://registry:5000"]`

	cfg := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{patch},
	}

	registries, err := docker.ExtractRegistriesFromKind(cfg)

	require.NoError(t, err)
	// Should deduplicate based on name:port
	assert.Len(t, registries, 1)
	assert.Equal(t, "registry", registries[0].Name)
	assert.Equal(t, "5000", registries[0].HostPort)
}

func TestExtractRegistriesFromKind_HTTPSEndpoint(t *testing.T) {
	t.Parallel()

	patch := `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
  endpoint = ["https://secure-registry:5000"]`

	assertKindSingleRegistry(t, patch, "secure-registry", "5000")
}

func TestExtractRegistriesFromKind_EndpointWithoutPort(t *testing.T) {
	t.Parallel()

	patch := `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
  endpoint = ["http://registry"]`

	assertKindSingleRegistry(t, patch, "registry", "5000")
}

func TestExtractRegistriesFromKind_EmptyPatch(t *testing.T) {
	t.Parallel()

	cfg := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{""},
	}

	registries, err := docker.ExtractRegistriesFromKind(cfg)

	require.NoError(t, err)
	assert.Empty(t, registries)
}

func TestExtractRegistriesFromKind_NilConfig(t *testing.T) {
	t.Parallel()

	registries, err := docker.ExtractRegistriesFromKind(nil)

	require.NoError(t, err)
	assert.Empty(t, registries)
}

func TestExtractRegistriesFromKind_ComplexContainerdPatch(t *testing.T) {
	t.Parallel()

	patch := `[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["https://registry-1.docker.io"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
      endpoint = ["http://k3d-registry:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs]
    [plugins."io.containerd.grpc.v1.cri".registry.configs."k3d-registry:5000".tls]
      insecure_skip_verify = true`

	cfg := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{patch},
	}

	registries, err := docker.ExtractRegistriesFromKind(cfg)

	require.NoError(t, err)
	// Should find the localhost:5000 mirror
	require.GreaterOrEqual(t, len(registries), 1)

	// Find the k3d-registry entry
	found := false

	for _, reg := range registries {
		if reg.Name == "k3d-registry" {
			found = true

			assert.Equal(t, "5000", reg.HostPort)

			break
		}
	}

	assert.True(t, found, "Should find k3d-registry in complex patch")
}
