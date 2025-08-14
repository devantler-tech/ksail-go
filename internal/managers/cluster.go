package managers

import (
	"fmt"

	factory "github.com/devantler-tech/ksail-go/internal/factories"
	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
)

// ClusterManager manages cluster operations like start, stop, and status checks.
type ClusterManager struct {
	config *ksailcluster.Cluster
}

// NewClusterManager creates a new ClusterManager instance.
func NewClusterManager(config *ksailcluster.Cluster) *ClusterManager {
	return &ClusterManager{config: config}
}

// OperationParams encapsulates the parameters for cluster operations.
type OperationParams struct {
	ActionMsg string
	VerbMsg   string
	PastMsg   string
}

// ExecuteOperation performs a cluster operation (start, stop, etc.) with the given parameters.
func (cm *ClusterManager) ExecuteOperation(params OperationParams, operation func(clusterprovisioner.ClusterProvisioner, string) error) error {
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
	fmt.Printf("%s '%s'\n", params.ActionMsg, cm.config.Metadata.Name)
	fmt.Printf("► checking '%s' is ready\n", cm.config.Spec.ContainerEngine)

	ready, err := containerEngineProvisioner.CheckReady()
	if err != nil || !ready {
		return fmt.Errorf("container engine '%s' is not ready: %v", cm.config.Spec.ContainerEngine, err)
	}

	fmt.Printf("✔ '%s' is ready\n", cm.config.Spec.ContainerEngine)
	fmt.Printf("► %s '%s'\n", params.VerbMsg, cm.config.Metadata.Name)

	exists, err := provisioner.Exists(cm.config.Metadata.Name)
	if err != nil {
		return err
	}

	if !exists {
		fmt.Printf("✔ '%s' not found\n", cm.config.Metadata.Name)
		return nil
	}

	if err := operation(provisioner, cm.config.Metadata.Name); err != nil {
		return err
	}

	fmt.Printf("✔ '%s' %s\n", cm.config.Metadata.Name, params.PastMsg)
	return nil
}