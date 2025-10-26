package kindprovisioner_test

import (
	"os"
	"path/filepath"
	"testing"

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
