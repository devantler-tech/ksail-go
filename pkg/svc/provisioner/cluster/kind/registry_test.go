package kindprovisioner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

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

// loadExpectedMap loads expected map results from JSON file.
func loadExpectedMap(t *testing.T, filename string) map[string][]string {
	t.Helper()
	//nolint:gosec // Test data files are safe
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to load expected data %s: %v", filename, err)
	}
	var result map[string][]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal expected data %s: %v", filename, err)
	}
	return result
}

func TestParseContainerdConfig(t *testing.T) {
	tests := []struct {
		name         string
		inputFile    string
		expectedFile string
	}{
		{
			name:         "standard single endpoint",
			inputFile:    "containerd_single_endpoint.toml",
			expectedFile: "expected_single_endpoint.json",
		},
		{
			name:         "multiple mirrors",
			inputFile:    "containerd_multiple_mirrors.toml",
			expectedFile: "expected_multiple_mirrors.json",
		},
		{
			name:         "multiple endpoints inline",
			inputFile:    "containerd_multiple_endpoints_inline.toml",
			expectedFile: "expected_multiple_endpoints.json",
		},
		{
			name:         "multiline array format",
			inputFile:    "containerd_multiline_array.toml",
			expectedFile: "expected_multiple_endpoints.json",
		},
		{
			name:         "extra whitespace",
			inputFile:    "containerd_extra_whitespace.toml",
			expectedFile: "expected_single_endpoint.json",
		},
		{
			name:         "with comments",
			inputFile:    "containerd_with_comments.toml",
			expectedFile: "expected_single_endpoint.json",
		},
		{
			name:         "registry with port and path",
			inputFile:    "containerd_registry_with_port.toml",
			expectedFile: "expected_registry_with_port.json",
		},
		{
			name:         "empty config",
			inputFile:    "containerd_empty.toml",
			expectedFile: "expected_empty.json",
		},
		{
			name:         "no endpoint field",
			inputFile:    "containerd_no_endpoint.toml",
			expectedFile: "expected_empty.json",
		},
		{
			name:         "malformed endpoint",
			inputFile:    "containerd_malformed.toml",
			expectedFile: "expected_empty.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			patch := loadTestData(t, tt.inputFile)
			expected := loadExpectedMap(t, tt.expectedFile)
			result := parseContainerdConfig(patch)
			assert.Equal(t, expected, result, "Parsed mirrors should match expected")
		})
	}
}

// loadExpectedRegistries loads expected RegistryInfo results from JSON file.
func loadExpectedRegistries(t *testing.T, filename string) []RegistryInfo {
	t.Helper()
	if filename == "" {
		return nil
	}
	//nolint:gosec // Test data files are safe
	data, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("failed to load expected data %s: %v", filename, err)
	}
	var result []RegistryInfo
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal expected data %s: %v", filename, err)
	}
	return result
}

func TestExtractRegistriesFromKind(t *testing.T) {
	tests := []struct {
		name         string
		inputFile    string
		expectedFile string
	}{
		{
			name:         "single registry",
			inputFile:    "containerd_single_endpoint.toml",
			expectedFile: "expected_registry_single.json",
		},
		{
			name:         "multiple registries",
			inputFile:    "containerd_multiple_mirrors.toml",
			expectedFile: "expected_registry_multiple.json",
		},
		{
			name:         "duplicate registries in multiple patches",
			inputFile:    "containerd_duplicate_patches.toml",
			expectedFile: "expected_registry_single.json",
		},
		{
			name:         "registry with special characters",
			inputFile:    "containerd_registry_special_chars.toml",
			expectedFile: "expected_registry_special_chars.json",
		},
		{
			name:         "no containerd patches",
			inputFile:    "",
			expectedFile: "",
		},
		{
			name:         "multiple endpoints uses first",
			inputFile:    "containerd_multiple_endpoints_inline.toml",
			expectedFile: "expected_registry_single.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var config *v1alpha4.Cluster
			if tt.inputFile == "" {
				config = &v1alpha4.Cluster{ContainerdConfigPatches: []string{}}
			} else {
				patch := loadTestData(t, tt.inputFile)
				config = &v1alpha4.Cluster{ContainerdConfigPatches: []string{patch}}
			}

			expected := loadExpectedRegistries(t, tt.expectedFile)
			result := extractRegistriesFromKind(config)
			assert.Equal(t, expected, result, "Extracted registries should match expected")
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
