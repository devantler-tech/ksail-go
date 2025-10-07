package cluster

import (
	"errors"
	"fmt"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewDeleteCmd creates and returns the delete command.
func NewDeleteCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Destroy a cluster",
		Long:         `Destroy a cluster.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = runtime.RunEWithRuntime(
		runtimeContainer,
		runtime.WithTimer(
			func(cmd *cobra.Command, injector runtime.Injector, tmr timer.Timer) error {
				factory, err := runtime.ResolveClusterProvisionerFactory(injector)
				if err != nil {
					return fmt.Errorf("resolve provisioner factory dependency: %w", err)
				}

				deps := DeleteDeps{
					Timer:   tmr,
					Factory: factory,
				}

				return HandleDeleteRunE(cmd, cfgManager, deps)
			},
		),
	)

	return cmd
}

// DeleteDeps contains the dependencies required to handle the delete command.
type DeleteDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

var errMissingClusterProvisionerForDelete = errors.New("missing cluster provisioner dependency")

// HandleDeleteRunE executes the cluster deletion workflow.
func HandleDeleteRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps DeleteDeps,
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
		return errMissingClusterProvisionerForDelete
	}

	clusterName, err := configmanager.GetClusterName(distributionConfig)
	if err != nil {
		return fmt.Errorf("failed to get cluster name from config: %w", err)
	}

	showDeletionTitle(cmd)

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "deleting cluster",
		Writer:  cmd.OutOrStdout(),
	})

	err = provisioner.Delete(cmd.Context(), clusterName)
	if err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "cluster deleted",
		Timer:      deps.Timer,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// showDeletionTitle displays the deletion stage title.
func showDeletionTitle(cmd *cobra.Command) {
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Delete cluster...",
		Emoji:   "🗑️",
		Writer:  cmd.OutOrStdout(),
	})
}
