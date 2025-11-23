package registry

import (
	"testing"
)

// TestIsLocalEndpointName tests endpoint locality detection.
func TestIsLocalEndpointName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "localhost exact", input: "localhost", expected: true},
		{name: "localhost uppercase", input: "LOCALHOST", expected: true},
		{name: "localhost with spaces", input: "  localhost  ", expected: true},
		{name: "0.0.0.0", input: "0.0.0.0", expected: true},
		{name: "127.0.0.1", input: "127.0.0.1", expected: true},
		{name: "127.x.x.x prefix", input: "127.1.2.3", expected: true},
		{name: "remote host", input: "registry.docker.io", expected: false},
		{name: "192.168.x.x", input: "192.168.1.1", expected: false},
		{name: "empty string", input: "", expected: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := isLocalEndpointName(testCase.input)
			if result != testCase.expected {
				t.Errorf(
					"isLocalEndpointName(%q) = %v, want %v",
					testCase.input,
					result,
					testCase.expected,
				)
			}
		})
	}
}

// TestExtractNameFromEndpoint tests hostname extraction from endpoints.
func TestExtractNameFromEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "http with port",
			endpoint: "http://registry.example.com:5000",
			expected: "registry.example.com",
		},
		{
			name:     "https with port",
			endpoint: "https://registry.example.com:5000",
			expected: "registry.example.com",
		},
		{
			name:     "http without port",
			endpoint: "http://registry.example.com",
			expected: "registry.example.com",
		},
		{name: "localhost with port", endpoint: "http://localhost:5000", expected: "localhost"},
		{name: "IP address with port", endpoint: "http://127.0.0.1:5000", expected: "127.0.0.1"},
		{name: "invalid - no protocol", endpoint: "registry.example.com:5000", expected: ""},
		{name: "invalid - empty string", endpoint: "", expected: ""},
		{name: "invalid - only protocol", endpoint: "http://", expected: ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractNameFromEndpoint(testCase.endpoint)
			if result != testCase.expected {
				t.Errorf(
					"ExtractNameFromEndpoint(%q) = %q, want %q",
					testCase.endpoint,
					result,
					testCase.expected,
				)
			}
		})
	}
}

// TestExtractPortFromEndpoint tests port extraction from endpoints.
func TestExtractPortFromEndpoint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		endpoint string
		expected int
	}{
		{
			name:     "valid http endpoint with port",
			endpoint: "http://registry.example.com:5000",
			expected: 5000,
		},
		{
			name:     "valid https endpoint with port",
			endpoint: "https://registry.example.com:5001",
			expected: 5001,
		},
		{name: "localhost with port", endpoint: "http://localhost:8080", expected: 8080},
		{name: "IP with port", endpoint: "http://127.0.0.1:3000", expected: 3000},
		{
			name:     "http without port returns 0",
			endpoint: "http://registry.example.com",
			expected: 0,
		},
		{
			name:     "https without port returns 0",
			endpoint: "https://registry.example.com",
			expected: 0,
		},
		{name: "invalid endpoint", endpoint: "not-a-url", expected: 0},
		{name: "invalid port number", endpoint: "http://registry.example.com:invalid", expected: 0},
		{name: "port out of range", endpoint: "http://registry.example.com:99999", expected: 0},
		{name: "empty endpoint", endpoint: "", expected: 0},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := ExtractPortFromEndpoint(testCase.endpoint)
			if result != testCase.expected {
				t.Errorf("ExtractPortFromEndpoint(%q) = %d, want %d",
					testCase.endpoint, result, testCase.expected)
			}
		})
	}
}

// TestExtractRegistryPort tests registry port extraction with port allocation.
func TestExtractRegistryPort(t *testing.T) {
	t.Parallel()

	t.Run("extracts port from first endpoint", func(t *testing.T) {
		t.Parallel()

		endpoints := []string{"http://localhost:5000"}
		usedPorts := make(map[int]struct{})
		nextPort := DefaultRegistryPort

		result := ExtractRegistryPort(endpoints, usedPorts, &nextPort)
		if result != DefaultRegistryPort {
			t.Errorf("expected %d, got %d", DefaultRegistryPort, result)
		}

		if _, exists := usedPorts[DefaultRegistryPort]; !exists {
			t.Errorf("expected port %d to be marked as used", DefaultRegistryPort)
		}

		if nextPort != DefaultRegistryPort+1 {
			t.Errorf("expected nextPort to be %d, got %d", DefaultRegistryPort+1, nextPort)
		}
	})

	t.Run("allocates next available port when no endpoint", func(t *testing.T) {
		t.Parallel()

		nextPort := ptrTo(DefaultRegistryPort)

		result := ExtractRegistryPort([]string{}, make(map[int]struct{}), nextPort)
		if result != DefaultRegistryPort {
			t.Errorf("expected %d, got %d", DefaultRegistryPort, result)
		}
	})

	t.Run("skips used ports", func(t *testing.T) {
		t.Parallel()

		usedPorts := map[int]struct{}{
			DefaultRegistryPort:     {},
			DefaultRegistryPort + 1: {},
		}
		result := ExtractRegistryPort([]string{}, usedPorts, ptrTo(DefaultRegistryPort))

		if result != DefaultRegistryPort+2 {
			t.Errorf("expected %d, got %d", DefaultRegistryPort+2, result)
		}
	})

	t.Run("uses default port when nextPort is nil", func(t *testing.T) {
		t.Parallel()

		result := ExtractRegistryPort([]string{}, make(map[int]struct{}), nil)
		if result != DefaultRegistryPort {
			t.Errorf("expected default port %d, got %d", DefaultRegistryPort, result)
		}
	})
}

func ptrTo(i int) *int {
	return &i
}
