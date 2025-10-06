package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/runtime"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// NewCreateCmd wires the cluster create command using the shared runtime container.
func NewCreateCmd(rt *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a cluster",
		Long:         `Create a Kubernetes cluster as defined by configuration.`,
		SilenceUsage: true,
	}

	selectors := []ksailconfigmanager.FieldSelector[v1alpha1.Cluster]{
		ksailconfigmanager.DefaultDistributionFieldSelector(),
		ksailconfigmanager.DefaultDistributionConfigFieldSelector(),
	}

	cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout(), selectors...)
	cfgManager.AddFlagsFromFields(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return rt.Invoke(cmd, func(injector do.Injector) error {
			tmr, err := do.Invoke[timer.Timer](injector)
			if err != nil {
				return fmt.Errorf("resolve timer dependency: %w", err)
			}

			factory, err := do.Invoke[clusterprovisioner.Factory](injector)
			if err != nil {
				return fmt.Errorf("resolve provisioner factory dependency: %w", err)
			}

			deps := CreateDeps{
				Timer:   tmr,
				Factory: factory,
			}

			return HandleCreateRunE(cmd, cfgManager, deps)
		})
	}

	return cmd
}

// CreateDeps contains the dependencies required to handle the create command.
type CreateDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

// HandleCreateRunE executes the cluster creation workflow.
func HandleCreateRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps CreateDeps,
) error {
	deps.Timer.Start()

	err := cfgManager.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	deps.Timer.NewStage()

	clusterCfg := cfgManager.GetConfig()
	provisioner, distributionConfig, err := deps.Factory.Create(cmd.Context(), clusterCfg)
	if err != nil {
		return fmt.Errorf("failed to resolve cluster provisioner: %w", err)
	}

	if provisioner == nil {
		return fmt.Errorf("missing cluster provisioner dependency")
	}

	clusterName, err := configmanager.GetClusterName(distributionConfig)
	if err != nil {
		return fmt.Errorf("failed to get cluster name from config: %w", err)
	}

	showProvisioningTitle(cmd)

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "creating cluster",
		Writer:  cmd.OutOrStdout(),
	})

	if err := provisioner.Create(cmd.Context(), clusterName); err != nil {
		return fmt.Errorf("failed to create cluster: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "cluster created",
		Timer:      deps.Timer,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// showProvisioningTitle displays the provisioning stage title.
func showProvisioningTitle(cmd *cobra.Command) {
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Create cluster...",
		Emoji:   "🚀",
		Writer:  cmd.OutOrStdout(),
	})
}
