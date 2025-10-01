package cluster

import (
	"context"
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	cmd := cmdhelpers.NewCobraCommand(
		"down",
		"Destroy the Kubernetes cluster",
		`Destroy the Kubernetes cluster and all of its associated resources.`,
		HandleDownRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)

	return cmd
}

// HandleDownRunE handles the down command.
// Exported for testing purposes.
func HandleDownRunE(
	_ *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	tmr := timer.New()
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
		return ErrUnsupportedDistribution
	}

	_, ferr := fmt.Fprintln(manager.Writer)
	if ferr != nil {
		return fmt.Errorf("write failure: %w", ferr)
	}

	notify.TitleMessage(manager.Writer, "🔥", notify.NewMessage("Destroying cluster..."))

	exists, err := provisioner.Exists(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to check cluster existence: %w", err)
	}

	if !exists {
		return ErrClusterNotFound
	}

	tmr.StartStage()
	notify.ActivityMessage(manager.Writer, notify.NewMessage("destroying cluster"))

	err = provisioner.Delete(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to destroy cluster: %w", err)
	}

	notify.SuccessMessage(
		manager.Writer,
		notify.NewMessage("cluster destroyed").WithTiming(tmr.Total(), tmr.Stage()),
	)

	return nil
}
