// Package cluster provides commands for managing Kubernetes clusters.
package cluster

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
)

const defaultStopTimeout = 5 * time.Minute

var (
	// ErrK3dNotImplemented is returned when K3d stop is not yet implemented.
	ErrK3dNotImplemented = errors.New("K3d stop is not yet implemented")
	// ErrEKSNotImplemented is returned when EKS stop is not yet implemented.
	ErrEKSNotImplemented = errors.New("EKS stop is not yet implemented")
	// ErrUnsupportedDistribution is returned when an unsupported distribution is specified.
	ErrUnsupportedDistribution = errors.New("unsupported distribution")
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		HandleStopRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}

// HandleStopRunE handles the stop command execution.
func HandleStopRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	// Load cluster configuration
	cluster, err := cmdhelpers.LoadClusterWithErrorHandling(cmd, manager)
	if err != nil {
		//nolint:wrapcheck // Error already wrapped and formatted by LoadClusterWithErrorHandling
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), defaultStopTimeout)
	defer cancel()

	// Create provisioner based on distribution
	provisioner, err := createProvisionerForStop(cluster)
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to create provisioner: "+err.Error())

		return fmt.Errorf("failed to create provisioner: %w", err)
	}

	// Stop the cluster
	clusterName := ""
	if cluster.Spec.Connection.Context != "" {
		clusterName = cluster.Spec.Connection.Context
	}

	err = provisioner.Stop(ctx, clusterName)
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to stop cluster: "+err.Error())

		return fmt.Errorf("failed to stop cluster: %w", err)
	}

	// Report success
	notify.Successln(cmd.OutOrStdout(), "Cluster stopped successfully")
	cmdhelpers.LogClusterInfo(cmd, []cmdhelpers.ClusterInfoField{
		{Label: "Distribution", Value: string(cluster.Spec.Distribution)},
		{Label: "Context", Value: cluster.Spec.Connection.Context},
	})

	return nil
}

// createProvisionerForStop creates a cluster provisioner for the stop operation.
// Currently only Kind is fully implemented.
//
//nolint:ireturn // Returning interface is the design pattern for factory functions
func createProvisionerForStop(
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, error) {
	switch cluster.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return createKindProvisionerForStop(cluster)
	case v1alpha1.DistributionK3d:
		return nil, ErrK3dNotImplemented
	case v1alpha1.DistributionEKS:
		return nil, ErrEKSNotImplemented
	case v1alpha1.DistributionTind:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDistribution, cluster.Spec.Distribution)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDistribution, cluster.Spec.Distribution)
	}
}

// createKindProvisionerForStop creates a Kind cluster provisioner.
//
//nolint:ireturn // Returning interface is the design pattern for factory functions
func createKindProvisionerForStop(
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, error) {
	// Load Kind configuration
	configPath := cluster.Spec.DistributionConfig
	if configPath == "" {
		configPath = "kind.yaml"
	}

	kindConfigMgr := kindconfigmanager.NewConfigManager(configPath)

	kindConfig, err := kindConfigMgr.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Kind config: %w", err)
	}

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Create Kind provider adapter
	kindProvider := &kindProviderAdapter{
		provider: kindcluster.NewProvider(),
	}

	// Create provisioner
	kubeConfig := cluster.Spec.Connection.Kubeconfig
	if kubeConfig == "" {
		kubeConfig = "~/.kube/config"
	}

	return kindprovisioner.NewKindClusterProvisioner(
		kindConfig,
		kubeConfig,
		kindProvider,
		dockerClient,
	), nil
}

// kindProviderAdapter adapts the real Kind provider to the KindProvider interface.
type kindProviderAdapter struct {
	provider *kindcluster.Provider
}

func (a *kindProviderAdapter) Create(name string, opts ...kindcluster.CreateOption) error {
	//nolint:wrapcheck // Direct passthrough to underlying provider
	return a.provider.Create(name, opts...)
}

func (a *kindProviderAdapter) Delete(name, kubeconfigPath string) error {
	//nolint:wrapcheck // Direct passthrough to underlying provider
	return a.provider.Delete(name, kubeconfigPath)
}

func (a *kindProviderAdapter) List() ([]string, error) {
	//nolint:wrapcheck // Direct passthrough to underlying provider
	return a.provider.List()
}

func (a *kindProviderAdapter) ListNodes(name string) ([]string, error) {
	nodes, err := a.provider.ListNodes(name)
	if err != nil {
		//nolint:wrapcheck // Error comes from underlying provider and doesn't need additional context
		return nil, err
	}
	// Convert nodes.Node to string names
	nodeNames := make([]string, len(nodes))
	for i, node := range nodes {
		nodeNames[i] = node.String()
	}

	return nodeNames, nil
}
