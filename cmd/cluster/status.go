package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"status",
		"Show status of the Kubernetes cluster",
		`Show the current status of the Kubernetes cluster.`,
		HandleStatusRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for status check operations",
			DefaultValue: metav1.Duration{Duration: defaultStatusTimeout},
		},
	)
}

// HandleStatusRunE handles the status command.
// Exported for testing purposes.
func HandleStatusRunE(
	_ *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	ctx := context.Background()

	config, err := manager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	engine, err := containerengine.GetAutoDetectedClient()
	if err != nil {
		return fmt.Errorf("failed to get container engine client: %w", err)
	}

	provisioner, err := newProvisioner(config, *engine)
	if err != nil {
		return fmt.Errorf("failed to create provisioner: %w", err)
	}

	distConfig, err := cmdhelpers.LoadDistributionConfig(config)
	if err != nil {
		return fmt.Errorf("failed to load distribution config: %w", err)
	}

	var clusterName string

	switch cfg := distConfig.(type) {
	case *v1alpha4.Cluster:
		clusterName = cfg.Name
	case *v1alpha5.SimpleConfig:
		clusterName = cfg.Name
	default:
		return fmt.Errorf("unsupported distribution config type")
	}

	fmt.Fprintln(manager.Writer)
	notify.TitleMessage(manager.Writer, "📊", notify.NewMessage("Checking cluster status..."))

	exists, err := provisioner.Exists(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to check cluster existence: %w", err)
	}

	if !exists {
		notify.ActivityMessage(manager.Writer, notify.NewMessage("cluster: not found"))

		return nil
	}

	notify.SuccessMessage(
		manager.Writer,
		notify.NewMessage("cluster: running"),
	)

	return nil
}
