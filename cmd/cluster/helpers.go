package cluster

import (
	"context"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
)

// createProvisionerForCluster creates a cluster provisioner based on the cluster configuration.
// If a provisioner factory is provided (for testing), it will be used instead.
// Returns the provisioner and cluster name.
//
//nolint:ireturn // Factory function returns interface for testability and flexibility.
func createProvisionerForCluster(
	ctx context.Context,
	cluster *v1alpha1.Cluster,
	provisioner provisionerFactory,
) (clusterprovisioner.ClusterProvisioner, string, error) {
	distribution := cluster.Spec.Distribution
	distributionConfigPath := cluster.Spec.DistributionConfig
	kubeconfigPath := cluster.Spec.Connection.Kubeconfig

	if provisioner != nil {
		return provisioner(
			ctx,
			distribution,
			distributionConfigPath,
			kubeconfigPath,
		)
	}

	// Load config once and get both provisioner and cluster name
	clusterProvisioner, clusterName, err := clusterprovisioner.CreateClusterProvisioner(
		ctx,
		distribution,
		distributionConfigPath,
		kubeconfigPath,
	)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create provisioner: %w", err)
	}

	return clusterProvisioner, clusterName, nil
}
