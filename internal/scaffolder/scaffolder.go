// Package scaffolder provides project scaffolding functionality for KSail.
package scaffolder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
)

const (
	// dirPerm is the permission for created directories.
	dirPerm = 0o755
	// filePerm is the permission for created files.
	filePerm = 0o644
)

// Scaffolder handles project scaffolding operations.
type Scaffolder struct {
	cluster v1alpha1.Cluster
}

// NewScaffolder creates a new Scaffolder instance.
func NewScaffolder(cluster v1alpha1.Cluster) *Scaffolder {
	return &Scaffolder{
		cluster: cluster,
	}
}

// Scaffold generates initial project files in the specified output directory.
func (s *Scaffolder) Scaffold(outputDir string, force bool) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, dirPerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate ksail.yaml configuration file
	if err := s.generateKSailConfig(outputDir, force); err != nil {
		return fmt.Errorf("failed to generate ksail.yaml: %w", err)
	}

	// Generate distribution-specific configuration file
	if err := s.generateDistributionConfig(outputDir, force); err != nil {
		return fmt.Errorf("failed to generate distribution config: %w", err)
	}

	// Generate k8s/kustomization.yaml entry point
	if err := s.generateKustomizationEntryPoint(outputDir, force); err != nil {
		return fmt.Errorf("failed to generate kustomization entry point: %w", err)
	}

	// Generate .sops.yaml for secret management (optional)
	if err := s.generateSopsConfig(outputDir, force); err != nil {
		return fmt.Errorf("failed to generate .sops.yaml: %w", err)
	}

	return nil
}

// generateKSailConfig generates the ksail.yaml configuration file.
func (s *Scaffolder) generateKSailConfig(outputDir string, force bool) error {
	generator := yamlgenerator.NewYAMLGenerator[v1alpha1.Cluster]()

	opts := yamlgenerator.Options{
		Output: filepath.Join(outputDir, "ksail.yaml"),
		Force:  force,
	}

	_, err := generator.Generate(s.cluster, opts)

	return err
}

// generateDistributionConfig generates the distribution-specific configuration file.
func (s *Scaffolder) generateDistributionConfig(outputDir string, force bool) error {
	configFile := s.cluster.Spec.DistributionConfig
	if configFile == "" {
		return nil // No distribution config specified
	}

	configPath := filepath.Join(outputDir, configFile)

	// Create a basic configuration based on the distribution
	var configContent string
	switch s.cluster.Spec.Distribution {
	case v1alpha1.DistributionKind:
		configContent = s.generateKindConfig()
	case v1alpha1.DistributionK3d:
		configContent = s.generateK3dConfig()
	case v1alpha1.DistributionEKS:
		configContent = s.generateEKSConfig()
	default:
		configContent = s.generateKindConfig() // Default to Kind
	}

	// Check if file exists and force is not set
	if !force {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("file %s already exists, use --force to overwrite", configPath)
		}
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(configPath), dirPerm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", configPath, err)
	}

	// Write the configuration file
	if err := os.WriteFile(configPath, []byte(configContent), filePerm); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}

	return nil
}

// generateKustomizationEntryPoint generates the k8s/kustomization.yaml entry point.
func (s *Scaffolder) generateKustomizationEntryPoint(outputDir string, force bool) error {
	sourceDir := s.cluster.Spec.SourceDirectory
	if sourceDir == "" {
		sourceDir = "k8s"
	}

	kustomizationDir := filepath.Join(outputDir, sourceDir)
	kustomizationPath := filepath.Join(kustomizationDir, "kustomization.yaml")

	// Check if file exists and force is not set
	if !force {
		if _, err := os.Stat(kustomizationPath); err == nil {
			return fmt.Errorf("file %s already exists, use --force to overwrite", kustomizationPath)
		}
	}

	// Create directory if needed
	if err := os.MkdirAll(kustomizationDir, dirPerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", kustomizationDir, err)
	}

	kustomizationContent := `apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources: []
`

	// Write the kustomization.yaml file
	if err := os.WriteFile(kustomizationPath, []byte(kustomizationContent), filePerm); err != nil {
		return fmt.Errorf("failed to write %s: %w", kustomizationPath, err)
	}

	return nil
}

// generateSopsConfig generates a basic .sops.yaml configuration.
func (s *Scaffolder) generateSopsConfig(outputDir string, force bool) error {
	sopsPath := filepath.Join(outputDir, ".sops.yaml")

	// Check if file exists and force is not set
	if !force {
		if _, err := os.Stat(sopsPath); err == nil {
			return nil // File exists and we don't want to overwrite
		}
	}

	sopsContent := `keys:
  - &default_key ""

creation_rules:
  - path_regex: \.enc\.ya?ml$
    encrypted_regex: ^(data|stringData)$
    pgp: *default_key
`

	// Write the .sops.yaml file
	if err := os.WriteFile(sopsPath, []byte(sopsContent), filePerm); err != nil {
		return fmt.Errorf("failed to write %s: %w", sopsPath, err)
	}

	return nil
}

// generateKindConfig returns a basic Kind cluster configuration.
func (s *Scaffolder) generateKindConfig() string {
	return `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ksail-default
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
`
}

// generateK3dConfig returns a basic K3d cluster configuration.
func (s *Scaffolder) generateK3dConfig() string {
	return `apiVersion: k3d.io/v1alpha4
kind: Simple
metadata:
  name: ksail-default
servers: 1
agents: 0
ports:
- port: 80:80
  nodeFilters:
  - loadbalancer
- port: 443:443
  nodeFilters:
  - loadbalancer
`
}

// generateEKSConfig returns a basic EKS cluster configuration.
func (s *Scaffolder) generateEKSConfig() string {
	return `apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig

metadata:
  name: ksail-default
  region: us-west-2

nodeGroups:
  - name: ng-1
    instanceType: m5.large
    desiredCapacity: 2
    minSize: 1
    maxSize: 4
`
}
