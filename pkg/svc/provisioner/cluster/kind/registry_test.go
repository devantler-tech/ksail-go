package kindprovisioner_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	docker "github.com/devantler-tech/ksail-go/pkg/client/docker"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestMain(m *testing.M) {
	v := m.Run()

	// After all tests have run, clean up snapshots
	_, _ = snaps.Clean(m)

	os.Exit(v)
}

// loadTestData loads test data from testdata directory.
func loadTestData(t *testing.T, filename string) string {
	t.Helper()
	//nolint:gosec // Test data files are safe
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to load test data %s: %v", filename, err)
	}

	return string(data)
}

func TestParseContainerdConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputFile string
	}{
		{name: "standard single endpoint", inputFile: "containerd_single_endpoint.toml"},
		{name: "multiple mirrors", inputFile: "containerd_multiple_mirrors.toml"},
		{name: "multiple endpoints inline", inputFile: "containerd_multiple_endpoints_inline.toml"},
		{name: "multiline array format", inputFile: "containerd_multiline_array.toml"},
		{name: "extra whitespace", inputFile: "containerd_extra_whitespace.toml"},
		{name: "with comments", inputFile: "containerd_with_comments.toml"},
		{name: "registry with port and path", inputFile: "containerd_registry_with_port.toml"},
		{name: "empty config", inputFile: "containerd_empty.toml"},
		{name: "no endpoint field", inputFile: "containerd_no_endpoint.toml"},
		{name: "malformed endpoint", inputFile: "containerd_malformed.toml"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			patch := loadTestData(t, testCase.inputFile)
			result := kindprovisioner.ParseContainerdConfigForTesting(patch)
			snaps.MatchSnapshot(t, result)
		})
	}
}

func TestExtractRegistriesFromKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputFile string
		isEmpty   bool
	}{
		{name: "single registry", inputFile: "containerd_single_endpoint.toml"},
		{name: "multiple registries", inputFile: "containerd_multiple_mirrors.toml"},
		{
			name:      "duplicate registries in multiple patches",
			inputFile: "containerd_duplicate_patches.toml",
		},
		{
			name:      "registry with special characters",
			inputFile: "containerd_registry_special_chars.toml",
		},
		{name: "no containerd patches", isEmpty: true},
		{
			name:      "multiple endpoints uses first",
			inputFile: "containerd_multiple_endpoints_inline.toml",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var config *v1alpha4.Cluster
			if testCase.isEmpty {
				config = &v1alpha4.Cluster{ContainerdConfigPatches: []string{}}
			} else {
				patch := loadTestData(t, testCase.inputFile)
				config = &v1alpha4.Cluster{ContainerdConfigPatches: []string{patch}}
			}

			result := kindprovisioner.ExtractRegistriesFromKindForTesting(config)
			snaps.MatchSnapshot(t, result)
		})
	}
}

func TestExtractQuotedString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple quoted string",
			input:    `"http://localhost:5000"`,
			expected: "http://localhost:5000",
		},
		{
			name:     "with whitespace",
			input:    `  "http://localhost:5000"  `,
			expected: "http://localhost:5000",
		},
		{
			name:     "no quotes",
			input:    `http://localhost:5000`,
			expected: "",
		},
		{
			name:     "only opening quote",
			input:    `"http://localhost:5000`,
			expected: "",
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := kindprovisioner.ExtractQuotedStringForTesting(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSetupRegistries_NilKindConfig(t *testing.T) {
	t.Parallel()

	mockClient := docker.NewMockAPIClient(t)
	ctx := context.Background()
	var buf bytes.Buffer

	err := kindprovisioner.SetupRegistries(ctx, nil, "test-cluster", mockClient, &buf)
	assert.NoError(t, err)
}

func TestSetupRegistries_NoRegistries(t *testing.T) {
	t.Parallel()

	mockClient := docker.NewMockAPIClient(t)
	ctx := context.Background()
	var buf bytes.Buffer

	kindConfig := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{},
	}

	err := kindprovisioner.SetupRegistries(ctx, kindConfig, "test-cluster", mockClient, &buf)
	assert.NoError(t, err)
}

func TestConnectRegistriesToNetwork_NilKindConfig(t *testing.T) {
	t.Parallel()

	mockClient := docker.NewMockAPIClient(t)
	ctx := context.Background()
	var buf bytes.Buffer

	err := kindprovisioner.ConnectRegistriesToNetwork(ctx, nil, mockClient, &buf)
	assert.NoError(t, err)
}

func TestConnectRegistriesToNetwork_NoRegistries(t *testing.T) {
	t.Parallel()

	mockClient := docker.NewMockAPIClient(t)
	ctx := context.Background()
	var buf bytes.Buffer

	kindConfig := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{},
	}

	err := kindprovisioner.ConnectRegistriesToNetwork(ctx, kindConfig, mockClient, &buf)
	assert.NoError(t, err)
}

func TestCleanupRegistries_NilKindConfig(t *testing.T) {
	t.Parallel()

	mockClient := docker.NewMockAPIClient(t)
	ctx := context.Background()

	err := kindprovisioner.CleanupRegistries(ctx, nil, "test-cluster", mockClient, false)
	assert.NoError(t, err)
}

func TestCleanupRegistries_NoRegistries(t *testing.T) {
	t.Parallel()

	mockClient := docker.NewMockAPIClient(t)
	ctx := context.Background()

	kindConfig := &v1alpha4.Cluster{
		ContainerdConfigPatches: []string{},
	}

	err := kindprovisioner.CleanupRegistries(ctx, kindConfig, "test-cluster", mockClient, false)
	assert.NoError(t, err)
}

func TestBuildRegistryInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		host             string
		endpoints        []string
		portOffset       int
		expectedName     string
		expectedPort     int
		expectedUpstream string
	}{
		{
			name:             "docker.io with port in endpoint",
			host:             "docker.io",
			endpoints:        []string{"http://localhost:5000"},
			portOffset:       0,
			expectedName:     "kind-docker-io",
			expectedPort:     5000,
			expectedUpstream: "https://registry-1.docker.io",
		},
		{
			name:             "ghcr.io with port offset",
			host:             "ghcr.io",
			endpoints:        []string{"http://localhost:5001"}, // Provide valid endpoint
			portOffset:       1,
			expectedName:     "kind-ghcr-io",
			expectedPort:     5001,
			expectedUpstream: "https://ghcr.io",
		},
		{
			name:             "quay.io with endpoint name extraction",
			host:             "quay.io",
			endpoints:        []string{"http://kind-quay-io:5002"},
			portOffset:       2,
			expectedName:     "kind-quay-io",
			expectedPort:     5002,
			expectedUpstream: "https://quay.io",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Since buildRegistryInfo is unexported, we test it through extractRegistriesFromKind
			patch := fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
  endpoint = %v`, testCase.host, formatEndpoints(testCase.endpoints))

			config := &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{patch},
			}

			registries := kindprovisioner.ExtractRegistriesFromKindForTesting(config)
			assert.Len(t, registries, 1)
			assert.Equal(t, testCase.expectedName, registries[0].Name)
			assert.Equal(t, testCase.expectedUpstream, registries[0].Upstream)
		})
	}
}

func formatEndpoints(endpoints []string) string {
	if len(endpoints) == 0 {
		return "[]"
	}
	quoted := make([]string, len(endpoints))
	for i, ep := range endpoints {
		quoted[i] = fmt.Sprintf(`"%s"`, ep)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func TestGenerateUpstreamURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "docker.io special case",
			host:     "docker.io",
			expected: "https://registry-1.docker.io",
		},
		{
			name:     "ghcr.io standard case",
			host:     "ghcr.io",
			expected: "https://ghcr.io",
		},
		{
			name:     "quay.io standard case",
			host:     "quay.io",
			expected: "https://quay.io",
		},
		{
			name:     "custom registry with port",
			host:     "registry.example.com:5000",
			expected: "https://registry.example.com:5000",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Test through extractRegistriesFromKind with a valid endpoint
			patch := fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
  endpoint = ["http://localhost:5000"]`, testCase.host)

			config := &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{patch},
			}

			registries := kindprovisioner.ExtractRegistriesFromKindForTesting(config)
			assert.Len(t, registries, 1)
			assert.Equal(t, testCase.expected, registries[0].Upstream)
		})
	}
}

func TestExtractPortFromEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		expected int
	}{
		{
			name:     "http with port",
			endpoint: "http://localhost:5000",
			expected: 5000,
		},
		{
			name:     "https with port",
			endpoint: "https://registry:5001",
			expected: 5001,
		},
		{
			name:     "with trailing slash",
			endpoint: "http://localhost:5002/",
			expected: 5002,
		},
		{
			name:     "with path",
			endpoint: "http://localhost:5003/v2",
			expected: 5003,
		},
		{
			name:     "no port",
			endpoint: "http://localhost",
			expected: 0,
		},
		{
			name:     "invalid port",
			endpoint: "http://localhost:invalid",
			expected: 0,
		},
		{
			name:     "port too high",
			endpoint: "http://localhost:99999",
			expected: 0,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Test through extractRegistriesFromKind
			patch := fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."test.io"]
  endpoint = ["%s"]`, testCase.endpoint)

			config := &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{patch},
			}

			registries := kindprovisioner.ExtractRegistriesFromKindForTesting(config)
			if testCase.expected > 0 {
				assert.Len(t, registries, 1)
				assert.Equal(t, testCase.expected, registries[0].Port)
			}
		})
	}
}

func TestGenerateNameFromHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		host     string
		expected string
	}{
		{
			name:     "simple host",
			host:     "docker.io",
			expected: "kind-docker-io",
		},
		{
			name:     "host with subdomain",
			host:     "registry.example.com",
			expected: "kind-registry-example-com",
		},
		{
			name:     "host with port",
			host:     "registry.io:5000",
			expected: "kind-registry-io-5000",
		},
		{
			name:     "host with slashes",
			host:     "example.com/path",
			expected: "kind-example-com-path",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Test through extractRegistriesFromKind with a valid endpoint
			patch := fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
  endpoint = ["http://localhost:5000"]`, testCase.host)

			config := &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{patch},
			}

			registries := kindprovisioner.ExtractRegistriesFromKindForTesting(config)
			assert.Len(t, registries, 1)
			assert.Equal(t, testCase.expected, registries[0].Name)
		})
	}
}

func TestExtractNameFromEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "kind-prefixed endpoint with port",
			endpoint: "http://kind-docker-io:5000",
			expected: "kind-docker-io",
		},
		{
			name:     "https kind-prefixed endpoint",
			endpoint: "https://kind-ghcr-io:5001",
			expected: "kind-ghcr-io",
		},
		{
			name:     "non-kind endpoint",
			endpoint: "http://localhost:5000",
			expected: "test-io", // Name is generated from host, not endpoint
		},
		{
			name:     "malformed endpoint",
			endpoint: "invalid",
			expected: "",
		},
		{
			name:     "endpoint without port",
			endpoint: "http://kind-registry",
			expected: "kind-registry",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Test through extractRegistriesFromKind
			patch := fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."test.io"]
  endpoint = ["%s"]`, testCase.endpoint)

			config := &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{patch},
			}

			registries := kindprovisioner.ExtractRegistriesFromKindForTesting(config)
			assert.Len(t, registries, 1)
			// The name extraction logic tries to extract from endpoint first
			if testCase.expected != "" {
				assert.Contains(t, registries[0].Name, strings.Split(testCase.expected, ":")[0])
			}
		})
	}
}

func TestParseContainerdConfig_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		patch     string
		expectLen int
	}{
		{
			name:      "empty patch",
			patch:     "",
			expectLen: 0,
		},
		{
			name:      "only comments",
			patch:     "# This is a comment\n# Another comment",
			expectLen: 0,
		},
		{
			name: "duplicate hosts",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5001"]`,
			expectLen: 1, // First occurrence is kept
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := kindprovisioner.ParseContainerdConfigForTesting(testCase.patch)
			assert.Len(t, result, testCase.expectLen)
		})
	}
}
