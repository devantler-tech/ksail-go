// Package eksprovisioner provides implementations of the Provisioner interface
// for provisioning EKS clusters using eksctl's Cobra commands.
package eksprovisioner

import (
	"context"
	"errors"
	"fmt"

	eksctlclient "github.com/devantler-tech/ksail-go/pkg/client/eksctl"
	ekstypes "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"k8s.io/cli-runtime/pkg/genericiooptions"
)

// ErrClusterNotFound is returned when a cluster is not found.
var ErrClusterNotFound = errors.New("cluster not found")

// ErrCommandCreationFailed is returned when eksctl command creation fails.
var ErrCommandCreationFailed = errors.New("failed to create eksctl command")

// EKSClusterProvisioner implements provisioning for EKS clusters using eksctl Cobra commands.
type EKSClusterProvisioner struct {
	clusterConfig *ekstypes.ClusterConfig
	client        *eksctlclient.Client
	configPath    string
}

// NewEKSClusterProvisioner constructs an EKS provisioner instance.
func NewEKSClusterProvisioner(
	clusterConfig *ekstypes.ClusterConfig,
	configPath string,
	client *eksctlclient.Client,
) *EKSClusterProvisioner {
	return &EKSClusterProvisioner{
		clusterConfig: clusterConfig,
		configPath:    configPath,
		client:        client,
	}
}

// NewDefaultEKSClusterProvisioner creates an EKS provisioner with default client.
func NewDefaultEKSClusterProvisioner(
	clusterConfig *ekstypes.ClusterConfig,
	configPath string,
) *EKSClusterProvisioner {
	ioStreams := genericiooptions.IOStreams{
		In:     nil,
		Out:    nil,
		ErrOut: nil,
	}
	client := eksctlclient.NewClient(ioStreams)

	return NewEKSClusterProvisioner(clusterConfig, configPath, client)
}

// Create provisions an EKS cluster using eksctl Cobra commands.
func (e *EKSClusterProvisioner) Create(_ context.Context, name string) error {
	target := name
	if target == "" && e.clusterConfig.Metadata != nil {
		target = e.clusterConfig.Metadata.Name
	}

	// Create the command
	cmd := e.client.CreateClusterCommand(e.configPath)
	if cmd == nil {
		return ErrCommandCreationFailed
	}

	// Build args
	args := []string{}
	if e.configPath != "" {
		args = append(args, "--config-file", e.configPath)
	}

	if target != "" {
		args = append(args, "--name", target)
	}

	// Execute the command
	_, err := e.client.ExecuteClusterCommand(cmd, args)
	if err != nil {
		return fmt.Errorf("failed to create EKS cluster: %w", err)
	}

	return nil
}

// Delete tears down an EKS cluster using eksctl Cobra commands.
func (e *EKSClusterProvisioner) Delete(_ context.Context, name string) error {
	target := name
	if target == "" && e.clusterConfig.Metadata != nil {
		target = e.clusterConfig.Metadata.Name
	}

	if target == "" {
		return ErrClusterNotFound
	}

	// Create the command
	cmd := e.client.DeleteClusterCommand(target)
	if cmd == nil {
		return ErrCommandCreationFailed
	}

	// Execute the command
	args := []string{"--name", target, "--wait"}

	_, err := e.client.ExecuteClusterCommand(cmd, args)
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

// List returns cluster names managed by eksctl using Cobra commands.
func (e *EKSClusterProvisioner) List(_ context.Context) ([]string, error) {
	clusters, err := e.client.ListClusters()
	if err != nil {
		return nil, fmt.Errorf("failed to list EKS clusters: %w", err)
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
