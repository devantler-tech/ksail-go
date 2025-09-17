package k3d

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	// No cleanup needed for these simple tests
	os.Exit(m.Run())
}

func TestNewConfigManager(t *testing.T) {
	filePath := "/test/path/config.yaml"
	manager := NewConfigManager(filePath)

	if manager == nil {
		t.Fatal("NewConfigManager returned nil")
	}

	if manager.filePath != filePath {
		t.Errorf("Expected filePath to be %s, got %s", filePath, manager.filePath)
	}

	if manager.config != nil {
		t.Error("Expected config to be nil initially")
	}

	if manager.configLoaded {
		t.Error("Expected configLoaded to be false initially")
	}
}

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		name        string
		configYAML  string
		expectError bool
		description string
	}{
		{
			name: "valid_simple_config",
			configYAML: `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: test-cluster
servers: 1
agents: 2
image: rancher/k3s:latest
network: test-network
`,
			expectError: false,
			description: "should load valid k3d simple config",
		},
		{
			name: "minimal_config",
			configYAML: `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: minimal-cluster
`,
			expectError: false,
			description: "should load minimal valid config",
		},
		{
			name: "invalid_yaml",
			configYAML: `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: test-cluster
  invalid: yaml: syntax
`,
			expectError: true,
			description: "should fail on invalid YAML syntax",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file with test config
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			err := os.WriteFile(configFile, []byte(tc.configYAML), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			// Create config manager and test LoadConfig
			manager := NewConfigManager(configFile)
			config, err := manager.LoadConfig()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none for %s", tc.description)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error for %s: %v", tc.description, err)
			}

			if config == nil {
				t.Fatal("LoadConfig returned nil config")
			}

			// Verify basic config structure for valid configs
			if config.APIVersion != "k3d.io/v1alpha5" {
				t.Errorf("Expected APIVersion 'k3d.io/v1alpha5', got '%s'", config.APIVersion)
			}

			if config.Kind != "Simple" {
				t.Errorf("Expected Kind 'Simple', got '%s'", config.Kind)
			}

			// Verify that subsequent calls return the same config (caching behavior)
			config2, err2 := manager.LoadConfig()
			if err2 != nil {
				t.Errorf("Second LoadConfig call failed: %v", err2)
			}

			if config != config2 {
				t.Error("Expected same config instance on subsequent calls")
			}
		})
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	nonExistentFile := "/non/existent/path/config.yaml"
	manager := NewConfigManager(nonExistentFile)

	config, err := manager.LoadConfig()

	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	if config != nil {
		t.Error("Expected nil config for failed load")
	}
}

func TestLoadConfigWithComplexConfig(t *testing.T) {
	complexConfig := `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: complex-cluster
servers: 3
agents: 5
image: rancher/k3s:v1.28.0-k3s1
network: custom-network
subnet: "172.20.0.0/16"
clusterToken: custom-token
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "complex-config.yaml")

	err := os.WriteFile(configFile, []byte(complexConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write complex config file: %v", err)
	}

	manager := NewConfigManager(configFile)
	config, err := manager.LoadConfig()

	if err != nil {
		t.Fatalf("Failed to load complex config: %v", err)
	}

	// Verify some complex fields
	if config.Name != "complex-cluster" {
		t.Errorf("Expected name 'complex-cluster', got '%s'", config.Name)
	}

	if config.Servers != 3 {
		t.Errorf("Expected 3 servers, got %d", config.Servers)
	}

	if config.Agents != 5 {
		t.Errorf("Expected 5 agents, got %d", config.Agents)
	}

	if config.Image != "rancher/k3s:v1.28.0-k3s1" {
		t.Errorf("Expected specific image, got '%s'", config.Image)
	}

	if config.Network != "custom-network" {
		t.Errorf("Expected custom-network, got '%s'", config.Network)
	}

	if config.ClusterToken != "custom-token" {
		t.Errorf("Expected custom-token, got '%s'", config.ClusterToken)
	}
}

func TestLoadConfigCreatesValidK3dTypes(t *testing.T) {
	simpleConfig := `apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: type-test-cluster
servers: 1
agents: 0
`

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "type-test.yaml")

	err := os.WriteFile(configFile, []byte(simpleConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write type test config file: %v", err)
	}

	manager := NewConfigManager(configFile)
	config, err := manager.LoadConfig()

	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify that the loaded config has the correct Go types
	if config.TypeMeta.APIVersion != "k3d.io/v1alpha5" {
		t.Errorf("TypeMeta.APIVersion incorrect: got '%s'", config.TypeMeta.APIVersion)
	}

	if config.TypeMeta.Kind != "Simple" {
		t.Errorf("TypeMeta.Kind incorrect: got '%s'", config.TypeMeta.Kind)
	}

	// Test type methods
	if config.GetAPIVersion() != "k3d.io/v1alpha5" {
		t.Errorf("GetAPIVersion() incorrect: got '%s'", config.GetAPIVersion())
	}

	if config.GetKind() != "Simple" {
		t.Errorf("GetKind() incorrect: got '%s'", config.GetKind())
	}

	// Verify ObjectMeta
	if config.ObjectMeta.Name != "type-test-cluster" {
		t.Errorf("ObjectMeta.Name incorrect: got '%s'", config.ObjectMeta.Name)
	}
}
