package registries_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeHostIdentifier(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"keepsAlpha", "dockerio", "dockerio"},
		{"keepsDots", "docker.io", "docker.io"},
		{"replacesSlashes", "ghcr.io/library/app", "ghcr.io-library-app"},
		{"replacesColons", "localhost:5000", "localhost-5000"},
		{"preservesWhitespace", "  example.com  ", "  example.com  "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, registries.SanitizeHostIdentifier(tc.input))
		})
	}
}

func TestGenerateUpstreamURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"dockerIOUsesMirror", "docker.io", "https://registry-1.docker.io"},
		{"keepsExistingScheme", "https://my-registry.local", "https://my-registry.local"},
		{"prefersHttps", "ghcr.io", "https://ghcr.io"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, registries.GenerateUpstreamURL(tc.input))
		})
	}
}

func TestExtractPortFromEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		endpoint string
		expected int
	}{
		{"withPort", "http://docker.io:5000", 5000},
		{"withPath", "https://mirror:5443/v2", 5443},
		{"missingPort", "https://mirror", 0},
		{"invalidPort", "https://mirror:notaport", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, registries.ExtractPortFromEndpoint(tc.endpoint))
		})
	}
}

func TestExtractRegistryPort_UsesEndpointPortWhenAvailable(t *testing.T) {
	t.Parallel()

	usedPorts := map[int]struct{}{}
	next := registries.DefaultRegistryPort

	port := registries.ExtractRegistryPort([]string{"http://ghcr.io:5050"}, usedPorts, &next)
	require.Equal(t, 5050, port)
	assert.Contains(t, usedPorts, 5050)
	assert.Equal(t, 5051, next)

	// Subsequent call with an already-used port falls back to next free port.
	port = registries.ExtractRegistryPort([]string{"http://ghcr.io:5050"}, usedPorts, &next)
	require.Equal(t, 5051, port)
	assert.Contains(t, usedPorts, 5051)
	assert.Equal(t, 5052, next)
}

func TestExtractRegistryPort_FallsBackToDefaultWhenNoEndpoint(t *testing.T) {
	t.Parallel()

	usedPorts := map[int]struct{}{registries.DefaultRegistryPort: {}}
	next := registries.DefaultRegistryPort

	port := registries.ExtractRegistryPort(nil, usedPorts, &next)
	require.Equal(t, registries.DefaultRegistryPort+1, port)
	assert.Contains(t, usedPorts, registries.DefaultRegistryPort+1)
	assert.Equal(t, registries.DefaultRegistryPort+2, next)

	// Nil next pointer should still allocate the default port when map is empty.
	newMap := map[int]struct{}{}
	port = registries.ExtractRegistryPort(nil, newMap, nil)
	require.Equal(t, registries.DefaultRegistryPort, port)
	assert.Contains(t, newMap, registries.DefaultRegistryPort)
}

func TestExtractNameFromEndpoint(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		endpoint string
		expected string
	}{
		{"valid", "http://docker.io:5000", "docker.io"},
		{"missingScheme", "docker.io:5000", ""},
		{"missingHost", "http://:5000", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, registries.ExtractNameFromEndpoint(tc.endpoint))
		})
	}
}

func TestResolveRegistryName(t *testing.T) {
	t.Parallel()

	t.Run("usesEndpointNameWhenPresent", func(t *testing.T) {
		t.Parallel()

		name := registries.ResolveRegistryName(
			"docker.io",
			[]string{"http://docker.io:5000"},
			"k3d-",
		)
		assert.Equal(t, "docker.io", name)
	})

	t.Run("fallsBackToPrefixAndHost", func(t *testing.T) {
		t.Parallel()

		name := registries.ResolveRegistryName(
			"ghcr.io",
			[]string{"invalid-endpoint"},
			"k3d-",
		)
		assert.Equal(t, "k3d-ghcr.io", name)
	})

	t.Run("ignoresLocalhostEndpoints", func(t *testing.T) {
		t.Parallel()

		name := registries.ResolveRegistryName(
			"docker.io",
			[]string{"http://localhost:5000"},
			"kind-",
		)
		assert.Equal(t, "kind-docker.io", name)
	})
}

func TestBuildRegistryInfo(t *testing.T) {
	t.Parallel()

	info := registries.BuildRegistryInfo(
		"docker.io",
		[]string{"http://docker.io:5000"},
		registries.DefaultRegistryPort,
		"",
		"",
	)

	require.Equal(t, "docker.io", info.Host)
	assert.Equal(t, "docker.io", info.Name)
	assert.Equal(t, "https://registry-1.docker.io", info.Upstream)
	assert.Equal(t, registries.DefaultRegistryPort, info.Port)
	assert.Equal(t, "docker.io", info.Volume)
}

func TestBuildRegistryInfo_UsesOverride(t *testing.T) {
	t.Parallel()

	info := registries.BuildRegistryInfo(
		"docker.io",
		[]string{"http://docker.io:5000"},
		registries.DefaultRegistryPort,
		"",
		"https://mirror.example.com",
	)

	require.Equal(t, "https://mirror.example.com", info.Upstream)
}

func TestBuildRegistryName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "k3d-ghcr.io", registries.BuildRegistryName("k3d-", "ghcr.io"))
}

func TestGenerateVolumeName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "ghcr.io", registries.GenerateVolumeName("ghcr.io"))
}

func TestSortHosts(t *testing.T) {
	t.Parallel()

	hosts := []string{"ghcr.io", "docker.io", "quay.io"}
	registries.SortHosts(hosts)
	assert.Equal(t, []string{"docker.io", "ghcr.io", "quay.io"}, hosts)
}
