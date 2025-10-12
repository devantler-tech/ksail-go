// Package eksprovisioner provides implementations of the Provisioner interface
// for provisioning EKS clusters using eksctl.
package eksprovisioner

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	ekstypes "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// ErrClusterNotFound is returned when a cluster is not found.
var ErrClusterNotFound = errors.New("cluster not found")

// EksctlExecutor defines the interface for executing eksctl commands.
type EksctlExecutor interface {
	Execute(ctx context.Context, args []string) (string, error)
}

// DefaultEksctlExecutor implements EksctlExecutor using os/exec.
type DefaultEksctlExecutor struct{}

// Execute runs eksctl with the provided arguments.
func (e *DefaultEksctlExecutor) Execute(ctx context.Context, args []string) (string, error) {
	cmd := exec.CommandContext(ctx, "eksctl", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("eksctl command failed: %w", err)
	}

	return string(output), nil
}

// EKSClusterProvisioner implements provisioning for EKS clusters using eksctl.
type EKSClusterProvisioner struct {
	clusterConfig *ekstypes.ClusterConfig
	executor      EksctlExecutor
	configPath    string
}

// NewEKSClusterProvisioner constructs an EKS provisioner instance.
func NewEKSClusterProvisioner(
	clusterConfig *ekstypes.ClusterConfig,
	configPath string,
	executor EksctlExecutor,
) *EKSClusterProvisioner {
	return &EKSClusterProvisioner{
		clusterConfig: clusterConfig,
		configPath:    configPath,
		executor:      executor,
	}
}

// NewDefaultEKSClusterProvisioner creates an EKS provisioner with default executor.
func NewDefaultEKSClusterProvisioner(
	clusterConfig *ekstypes.ClusterConfig,
	configPath string,
) *EKSClusterProvisioner {
	return NewEKSClusterProvisioner(clusterConfig, configPath, &DefaultEksctlExecutor{})
}

// Create provisions an EKS cluster using eksctl.
func (e *EKSClusterProvisioner) Create(ctx context.Context, name string) error {
	target := name
	if target == "" && e.clusterConfig.Metadata != nil {
		target = e.clusterConfig.Metadata.Name
	}

	args := []string{"create", "cluster"}
	if e.configPath != "" {
		args = append(args, "--config-file", e.configPath)
	}

	if target != "" {
		args = append(args, "--name", target)
	}

	_, err := e.executor.Execute(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to create EKS cluster: %w", err)
	}

	return nil
}

// Delete tears down an EKS cluster.
func (e *EKSClusterProvisioner) Delete(ctx context.Context, name string) error {
	target := name
	if target == "" && e.clusterConfig.Metadata != nil {
		target = e.clusterConfig.Metadata.Name
	}

	if target == "" {
		return ErrClusterNotFound
	}

	args := []string{"delete", "cluster", "--name", target, "--wait"}

	_, err := e.executor.Execute(ctx, args)
	if err != nil {
		return fmt.Errorf("failed to delete EKS cluster: %w", err)
	}

	return nil
}

// Start starts an existing EKS cluster.
// Note: EKS clusters are always running once created. This is a no-op for compatibility.
func (e *EKSClusterProvisioner) Start(_ context.Context, _ string) error {
	// EKS clusters don't have a "stopped" state - they're always running
	return nil
}

// Stop stops a running EKS cluster.
// Note: EKS clusters cannot be stopped, only deleted. This is a no-op for compatibility.
func (e *EKSClusterProvisioner) Stop(_ context.Context, _ string) error {
	// EKS clusters don't support stop operation
	return nil
}

// List returns cluster names managed by eksctl.
func (e *EKSClusterProvisioner) List(ctx context.Context) ([]string, error) {
	args := []string{"get", "cluster", "-o", "json"}

	output, err := e.executor.Execute(ctx, args)
	if err != nil {
		// If no clusters exist, eksctl may return an error
		if strings.Contains(output, "No cluster found") ||
			strings.Contains(err.Error(), "No cluster found") {
			return []string{}, nil
		}

		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
	}

	// Parse cluster names from output
	// For simplicity, we'll parse the JSON output for cluster names
	// This is a basic implementation - production would use proper JSON parsing
	lines := strings.Split(output, "\n")

	var clusters []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "\"name\":") {
			// Extract cluster name from JSON field
			parts := strings.Split(line, ":")
			minParts := 2

			if len(parts) >= minParts {
				name := strings.Trim(strings.TrimSuffix(parts[1], ","), "\" ")
				if name != "" {
					clusters = append(clusters, name)
				}
			}
		}
	}

	return clusters, nil
}

// Exists checks if an EKS cluster exists.
func (e *EKSClusterProvisioner) Exists(ctx context.Context, name string) (bool, error) {
	clusters, err := e.List(ctx)
	if err != nil {
		return false, err
	}

	target := name
	if target == "" && e.clusterConfig.Metadata != nil {
		target = e.clusterConfig.Metadata.Name
	}

	for _, cluster := range clusters {
		if cluster == target {
			return true, nil
		}
	}

	return false, nil
}
