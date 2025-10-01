package cluster

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	k3dconfig "github.com/devantler-tech/ksail-go/pkg/config-manager/k3d"
	kindconfig "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/cluster"
)

const defaultUpTimeout = 5 * time.Minute

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	cmd := cmdhelpers.NewCobraCommand(
		"up",
		"Start the Kubernetes cluster",
		`Start the Kubernetes cluster defined in the project configuration.`,
		HandleUpRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for cluster operations",
			DefaultValue: metav1.Duration{Duration: defaultUpTimeout},
		},
	)

	// Add force flag for delete/recreate semantics
	cmd.Flags().Bool("force", false, "Force delete and recreate cluster if it already exists")

	return cmd
}

// HandleUpRunE handles the up command.
// Exported for testing purposes.
func HandleUpRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	tmr := timer.New()
	ctx := context.Background()

	config, err := manager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}
	force, _ := cmd.Flags().GetBool("force")

	engine, err := containerengine.GetAutoDetectedClient()
	if err != nil {
		return fmt.Errorf("failed to get container engine client: %w", err)
	}

	provisioner, err := newProvisioner(config, *engine)
	if err != nil {
		return fmt.Errorf("failed to create provisioner: %w", err)
	}

	fmt.Fprintln(manager.Writer)
	notify.TitleMessage(manager.Writer, "🚀", notify.NewMessage("Provisioning cluster..."))
	if err := provisionCluster(ctx, manager.Writer, tmr, provisioner, force); err != nil {
		return err
	}

	return nil
}

// newProvisioner creates and wires the cluster provisioner based on distribution
func newProvisioner(
	config *v1alpha1.Cluster,
	engine containerengine.ContainerEngine,
) (clusterprovisioner.ClusterProvisioner, error) {
	switch config.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return newKindProvisioner(config, engine)
	case v1alpha1.DistributionK3d:
		return newK3dProvisioner(config)
	default:
		return nil, fmt.Errorf("unsupported distribution: %s", config.Spec.Distribution)
	}
}

// newKindProvisioner creates a Kind cluster provisioner
func newKindProvisioner(
	config *v1alpha1.Cluster,
	engine containerengine.ContainerEngine,
) (clusterprovisioner.ClusterProvisioner, error) {
	kindConfigManager := kindconfig.NewConfigManager(config.Spec.DistributionConfig, os.Stdout)
	kindConfig, err := kindConfigManager.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Kind configuration: %w", err)
	}

	provider := kindprovisioner.NewKindProviderAdapter(cluster.NewProvider())

	return kindprovisioner.NewKindClusterProvisioner(
		kindConfig,
		config.Spec.Connection.Kubeconfig,
		provider,
		engine.Client,
	), nil
}

// newK3dProvisioner creates a K3d cluster provisioner
func newK3dProvisioner(config *v1alpha1.Cluster) (clusterprovisioner.ClusterProvisioner, error) {
	k3dConfigManager := k3dconfig.NewConfigManager(config.Spec.DistributionConfig, os.Stdout)
	k3dConfig, err := k3dConfigManager.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load K3d configuration: %w", err)
	}

	clientAdapter := k3dprovisioner.NewK3dClientAdapter()
	configAdapter := k3dprovisioner.NewK3dConfigAdapter()

	return k3dprovisioner.NewK3dClusterProvisioner(
		k3dConfig,
		clientAdapter,
		configAdapter,
	), nil
}

// checkPrerequisites verifies prerequisites are met for the selected distribution
func provisionCluster(
	ctx context.Context,
	writer io.Writer,
	tmr *timer.Timer,
	provisioner clusterprovisioner.ClusterProvisioner,
	force bool,
) error {
	exists, err := provisioner.Exists(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to check cluster existence: %w", err)
	}

	if exists {
		if !force {
			return fmt.Errorf("cluster already exists (use --force to recreate)")
		}

		if err := forceRecreateCluster(ctx, writer, tmr, provisioner); err != nil {
			return err
		}

		exists = false
	}

	if !exists {
		return createCluster(ctx, writer, tmr, provisioner)
	}

	return nil
}

// forceRecreateCluster deletes an existing cluster for force recreation.
func forceRecreateCluster(
	ctx context.Context,
	writer io.Writer,
	tmr *timer.Timer,
	provisioner clusterprovisioner.ClusterProvisioner,
) error {
	tmr.StartStage()
	notify.ActivityMessage(writer, notify.NewMessage("destroying existing cluster"))

	if err := provisioner.Delete(ctx, ""); err != nil {
		return fmt.Errorf("failed to destroy cluster: %w", err)
	}

	notify.SuccessMessage(
		writer,
		notify.NewMessage("existing cluster destroyed").WithTiming(tmr.Total(), tmr.Stage()),
	)
	return nil
}

// createCluster provisions a new cluster.
func createCluster(
	ctx context.Context,
	writer io.Writer,
	tmr *timer.Timer,
	provisioner clusterprovisioner.ClusterProvisioner,
) error {
	tmr.StartStage()
	notify.ActivityMessage(writer, notify.NewMessage("creating cluster"))

	if err := provisioner.Create(ctx, ""); err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	notify.SuccessMessage(
		writer,
		notify.NewMessage("cluster created").WithTiming(tmr.Total(), tmr.Stage()),
	)
	return nil
}
