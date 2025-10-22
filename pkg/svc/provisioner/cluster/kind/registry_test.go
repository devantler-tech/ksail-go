package kindprovisioner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestParseContainerdConfig(t *testing.T) {
	tests := []struct {
		name     string
		patch    string
		expected map[string][]string
	}{
		{
			name: "standard single endpoint",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000"]`,
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
			},
		},
		{
			name: "multiple mirrors",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."gcr.io"]
  endpoint = ["http://localhost:5001"]`,
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
				"gcr.io":    {"http://localhost:5001"},
			},
		},
		{
			name: "multiple endpoints inline",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000", "http://localhost:5001"]`,
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000", "http://localhost:5001"},
			},
		},
		{
			name: "multiline array format",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = [
    "http://localhost:5000",
    "http://localhost:5001"
  ]`,
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000", "http://localhost:5001"},
			},
		},
		{
			name: "extra whitespace",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
    endpoint   =   [  "http://localhost:5000"  ]`,
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
			},
		},
		{
			name: "with comments",
			patch: `# Mirror configuration
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  # Primary endpoint
  endpoint = ["http://localhost:5000"]`,
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
			},
		},
		{
			name: "registry with port and path",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry.example.com:5000"]
  endpoint = ["http://mirror.example.com:8080/v2"]`,
			expected: map[string][]string{
				"registry.example.com:5000": {"http://mirror.example.com:8080/v2"},
			},
		},
		{
			name:     "empty config",
			patch:    ``,
			expected: map[string][]string{},
		},
		{
			name: "no endpoint field",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  # Missing endpoint`,
			expected: map[string][]string{},
		},
		{
			name: "malformed endpoint",
			patch: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = [not-a-valid-url]`,
			expected: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseContainerdConfig(tt.patch)
			assert.Equal(t, tt.expected, result, "Parsed mirrors should match expected")
		})
	}
}

func TestExtractRegistriesFromKind(t *testing.T) {
	tests := []struct {
		name     string
		config   *v1alpha4.Cluster
		expected []RegistryInfo
	}{
		{
			name: "single registry",
			config: &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{
					`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000"]`,
				},
			},
			expected: []RegistryInfo{
				{
					Name:     "docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
			},
		},
		{
			name: "multiple registries",
			config: &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{
					`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000"]
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."gcr.io"]
  endpoint = ["http://localhost:5001"]`,
				},
			},
			expected: []RegistryInfo{
				{
					Name:     "docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
				{
					Name:     "gcr-io",
					Upstream: "https://gcr.io",
					Port:     5001,
				},
			},
		},
		{
			name: "duplicate registries in multiple patches",
			config: &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{
					`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000"]`,
					`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5001"]`,
				},
			},
			expected: []RegistryInfo{
				{
					Name:     "docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
			},
		},
		{
			name: "registry with special characters",
			config: &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{
					`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."registry.example.com:5000/path"]
  endpoint = ["http://mirror.example.com"]`,
				},
			},
			expected: []RegistryInfo{
				{
					Name:     "registry-example-com-5000-path",
					Upstream: "https://registry.example.com:5000/path",
					Port:     5000,
				},
			},
		},
		{
			name: "no containerd patches",
			config: &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{},
			},
			expected: nil,
		},
		{
			name: "multiple endpoints uses first",
			config: &v1alpha4.Cluster{
				ContainerdConfigPatches: []string{
					`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
  endpoint = ["http://localhost:5000", "http://localhost:5001"]`,
				},
			},
			expected: []RegistryInfo{
				{
					Name:     "docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRegistriesFromKind(tt.config)
			assert.Equal(t, tt.expected, result, "Extracted registries should match expected")
		})
	}
}

func TestExtractQuotedString(t *testing.T) {
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
			result := extractQuotedString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
