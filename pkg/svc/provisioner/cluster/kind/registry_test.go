package kindprovisioner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// loadTestData loads test data from testdata directory.
func loadTestData(t *testing.T, filename string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to load test data %s: %v", filename, err)
	}
	return string(data)
}

func TestParseContainerdConfig(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		expected map[string][]string
	}{
		{
			name: "standard single endpoint",
			file: "containerd_single_endpoint.toml",
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
			},
		},
		{
			name: "multiple mirrors",
			file: "containerd_multiple_mirrors.toml",
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
				"gcr.io":    {"http://localhost:5001"},
			},
		},
		{
			name: "multiple endpoints inline",
			file: "containerd_multiple_endpoints_inline.toml",
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000", "http://localhost:5001"},
			},
		},
		{
			name: "multiline array format",
			file: "containerd_multiline_array.toml",
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000", "http://localhost:5001"},
			},
		},
		{
			name: "extra whitespace",
			file: "containerd_extra_whitespace.toml",
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
			},
		},
		{
			name: "with comments",
			file: "containerd_with_comments.toml",
			expected: map[string][]string{
				"docker.io": {"http://localhost:5000"},
			},
		},
		{
			name: "registry with port and path",
			file: "containerd_registry_with_port.toml",
			expected: map[string][]string{
				"registry.example.com:5000": {"http://mirror.example.com:8080/v2"},
			},
		},
		{
			name:     "empty config",
			file:     "containerd_empty.toml",
			expected: map[string][]string{},
		},
		{
			name:     "no endpoint field",
			file:     "containerd_no_endpoint.toml",
			expected: map[string][]string{},
		},
		{
			name:     "malformed endpoint",
			file:     "containerd_malformed.toml",
			expected: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			patch := loadTestData(t, tt.file)
			result := parseContainerdConfig(patch)
			assert.Equal(t, tt.expected, result, "Parsed mirrors should match expected")
		})
	}
}

func TestExtractRegistriesFromKind(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		expected []RegistryInfo
	}{
		{
			name: "single registry",
			file: "containerd_single_endpoint.toml",
			expected: []RegistryInfo{
				{
					Name:     "kind-docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
			},
		},
		{
			name: "multiple registries",
			file: "containerd_multiple_mirrors.toml",
			expected: []RegistryInfo{
				{
					Name:     "kind-docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
				{
					Name:     "kind-gcr-io",
					Upstream: "https://gcr.io",
					Port:     5001,
				},
			},
		},
		{
			name: "duplicate registries in multiple patches",
			file: "containerd_duplicate_patches.toml",
			expected: []RegistryInfo{
				{
					Name:     "kind-docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
			},
		},
		{
			name: "registry with special characters",
			file: "containerd_registry_special_chars.toml",
			expected: []RegistryInfo{
				{
					Name:     "kind-registry-example-com-5000-path",
					Upstream: "https://registry.example.com:5000/path",
					Port:     5000,
				},
			},
		},
		{
			name:     "no containerd patches",
			file:     "",
			expected: nil,
		},
		{
			name: "multiple endpoints uses first",
			file: "containerd_multiple_endpoints_inline.toml",
			expected: []RegistryInfo{
				{
					Name:     "kind-docker-io",
					Upstream: "https://registry-1.docker.io",
					Port:     5000,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var config *v1alpha4.Cluster
			if tt.file == "" {
				config = &v1alpha4.Cluster{ContainerdConfigPatches: []string{}}
			} else {
				patch := loadTestData(t, tt.file)
				config = &v1alpha4.Cluster{ContainerdConfigPatches: []string{patch}}
			}

			result := extractRegistriesFromKind(config)
			assert.Equal(t, tt.expected, result, "Extracted registries should match expected")
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

			result := extractQuotedString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
