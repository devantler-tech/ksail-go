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

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		HandleStopRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}

// HandleStopRunE handles the stop command.
// Exported for testing purposes.
func HandleStopRunE(
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
	notify.TitleMessage(manager.Writer, "⏸️", notify.NewMessage("Stopping cluster..."))

	exists, err := provisioner.Exists(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to check cluster existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("cluster does not exist")
	}

	tmr.StartStage()
	notify.ActivityMessage(manager.Writer, notify.NewMessage("stopping cluster"))

	if err := provisioner.Stop(ctx, clusterName); err != nil {
		return fmt.Errorf("failed to stop cluster: %w", err)
	}

	notify.SuccessMessage(
		manager.Writer,
		notify.NewMessage("cluster stopped").WithTiming(tmr.Total(), tmr.Stage()),
	)

	return nil
}
