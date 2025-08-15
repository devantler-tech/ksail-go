package managers

import (
	"fmt"

	factory "github.com/devantler-tech/ksail-go/internal/factories"
	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
)

// ClusterOperation represents the type of cluster operation to perform.
type ClusterOperation int

const (
	// Start represents starting a cluster.
	Start ClusterOperation = iota
	// Stop represents stopping a cluster.
	Stop
)

// ClusterManager manages cluster operations like start, stop, and status checks.
type ClusterManager struct {
	config *ksailcluster.Cluster
}

// NewClusterManager creates a new ClusterManager instance.
func NewClusterManager(config *ksailcluster.Cluster) *ClusterManager {
	return &ClusterManager{config: config}
}

// StartOrStopCluster starts or stops the cluster based on the operation type.
func (cm *ClusterManager) StartOrStopCluster(operation ClusterOperation) error {
	// Derive messages and operation from operation type
	actionMsg, verbMsg, pastMsg, provisionerOp, err := cm.getOperationDetails(operation)
	if err != nil {
		return err
	}

	return cm.executeOperation(actionMsg, verbMsg, pastMsg, provisionerOp)
}

// getOperationDetails returns the operation details based on the operation type.
func (cm *ClusterManager) getOperationDetails(operation ClusterOperation) (string, string, string, func(clusterprovisioner.ClusterProvisioner, string) error, error) {
	switch operation {
	case Start:
		return "▶️ Starting", "starting", "started", func(p clusterprovisioner.ClusterProvisioner, name string) error {
			return p.Start(name)
		}, nil
	case Stop:
		return "⏹️ Stopping", "stopping", "stopped", func(p clusterprovisioner.ClusterProvisioner, name string) error {
			return p.Stop(name)
		}, nil
	default:
		return "", "", "", nil, fmt.Errorf("unsupported operation: %d", operation)
	}
}

// executeOperation executes the cluster operation with the given parameters.
func (cm *ClusterManager) executeOperation(actionMsg, verbMsg, pastMsg string, provisionerOp func(clusterprovisioner.ClusterProvisioner, string) error) error {
	fmt.Println()

	provisioner, err := factory.ClusterProvisioner(cm.config)
	if err != nil {
		return err
	}

	containerEngineProvisioner, err := factory.ContainerEngineProvisioner(cm.config)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("%s '%s'\n", actionMsg, cm.config.Metadata.Name)
	fmt.Printf("► checking '%s' is ready\n", cm.config.Spec.ContainerEngine)

	ready, err := containerEngineProvisioner.CheckReady()
	if err != nil || !ready {
		return fmt.Errorf("container engine '%s' is not ready: %v", cm.config.Spec.ContainerEngine, err)
	}

	fmt.Printf("✔ '%s' is ready\n", cm.config.Spec.ContainerEngine)
	fmt.Printf("► %s '%s'\n", verbMsg, cm.config.Metadata.Name)

	exists, err := provisioner.Exists(cm.config.Metadata.Name)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("✔ '%s' not found\n", cm.config.Metadata.Name)
		return nil
	}

	if err := provisionerOp(provisioner, cm.config.Metadata.Name); err != nil {
		return err
	}

	fmt.Printf("✔ '%s' %s\n", cm.config.Metadata.Name, pastMsg)

	return nil
}