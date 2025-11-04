package docker

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

// TestIsClusterNetworkName tests all branches of the network name validation.
func TestIsClusterNetworkName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		network  string
		expected bool
	}{
		{name: "empty string", network: "", expected: false},
		{name: "kind exact match", network: "kind", expected: true},
		{name: "kind with prefix", network: "kind-cluster", expected: true},
		{name: "k3d exact match", network: "k3d", expected: true},
		{name: "k3d with prefix", network: "k3d-cluster", expected: true},
		{name: "invalid network", network: "bridge", expected: false},
		{name: "partial match", network: "kindred", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isClusterNetworkName(tc.network)
			if result != tc.expected {
				t.Errorf("isClusterNetworkName(%q) = %v, want %v", tc.network, result, tc.expected)
			}
		})
	}
}

// TestDeriveRegistryVolumeName tests volume name derivation logic.
func TestDeriveRegistryVolumeName(t *testing.T) {
	t.Parallel()

	t.Run("extracts volume from mounts", func(t *testing.T) {
		t.Parallel()

		registry := container.Summary{
			Mounts: []types.MountPoint{
				{
					Type: mount.TypeVolume,
					Name: "registry-volume",
				},
			},
		}

		result := deriveRegistryVolumeName(registry, "fallback")
		if result != "registry-volume" {
			t.Errorf("expected 'registry-volume', got %q", result)
		}
	})

	t.Run("uses normalized fallback when no volume mounts", func(t *testing.T) {
		t.Parallel()

		registry := container.Summary{
			Mounts: []types.MountPoint{
				{
					Type: mount.TypeBind,
					Name: "",
				},
			},
		}

		// NormalizeVolumeName with kind- prefix removes it
		result := deriveRegistryVolumeName(registry, "kind-fallback")
		expected := "fallback"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})

	t.Run("returns trimmed fallback when normalization returns empty", func(t *testing.T) {
		t.Parallel()

		registry := container.Summary{
			Mounts: []types.MountPoint{},
		}

		result := deriveRegistryVolumeName(registry, "  spaces  ")
		expected := "spaces"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}

// TestNormalizeVolumeName tests volume name normalization.
func TestNormalizeVolumeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple name", input: "registry", expected: "registry"},
		{name: "with kind prefix", input: "kind-registry", expected: "registry"},
		{name: "with k3d prefix", input: "k3d-registry", expected: "registry"},
		{name: "kind prefix only", input: "kind-", expected: "kind-"},
		{name: "k3d prefix only", input: "k3d-", expected: "k3d-"},
		{name: "with spaces", input: "  test  ", expected: "test"},
		{name: "empty after trim", input: "   ", expected: ""},
		{name: "no prefix removal for slashes", input: "test/registry", expected: "test/registry"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := NormalizeVolumeName(tc.input)
			if result != tc.expected {
				t.Errorf("NormalizeVolumeName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// TestResolveVolumeName tests volume name resolution logic.
func TestResolveVolumeName(t *testing.T) {
	t.Parallel()

	// Create a mock client and registry manager
	mockClient := NewMockAPIClient(t)
	manager, err := NewRegistryManager(mockClient)
	if err != nil {
		t.Fatalf("failed to create registry manager: %v", err)
	}

	t.Run("uses provided volume name", func(t *testing.T) {
		t.Parallel()

		config := RegistryConfig{
			Name:       "test",
			VolumeName: "custom-volume",
		}
		result := manager.resolveVolumeName(config)
		if result != "custom-volume" {
			t.Errorf("expected 'custom-volume', got %q", result)
		}
	})

	t.Run("uses normalized name when volume name is empty", func(t *testing.T) {
		t.Parallel()

		config := RegistryConfig{
			Name:       "test-registry",
			VolumeName: "",
		}
		result := manager.resolveVolumeName(config)
		if result != "test-registry" {
			t.Errorf("expected 'test-registry', got %q", result)
		}
	})
}
