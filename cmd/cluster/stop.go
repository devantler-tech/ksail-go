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

// NewStopCmd creates and returns the stop command.
func NewStopCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "stop",
		Short:        "Stop a running cluster",
		Long:         `Stop a running Kubernetes cluster.`,
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

				deps := StopDeps{
					Timer:   tmr,
					Factory: factory,
				}

				return HandleStopRunE(cmd, cfgManager, deps)
			},
		),
	)

	return cmd
}

// StopDeps contains the dependencies required to handle the stop command.
type StopDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

var errMissingClusterProvisionerForStop = errors.New("missing cluster provisioner dependency")

// HandleStopRunE executes the cluster stop workflow.
func HandleStopRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps StopDeps,
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
		return errMissingClusterProvisionerForStop
	}

	clusterName, err := configmanager.GetClusterName(distributionConfig)
	if err != nil {
		return fmt.Errorf("failed to get cluster name: %w", err)
	}

	showStoppingTitle(cmd)

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "stopping cluster",
		Writer:  cmd.OutOrStdout(),
	})

	err = provisioner.Stop(cmd.Context(), clusterName)
	if err != nil {
		return fmt.Errorf("failed to stop cluster: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "cluster stopped",
		Timer:      deps.Timer,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// showStoppingTitle displays the stopping stage title.
func showStoppingTitle(cmd *cobra.Command) {
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Stop cluster...",
		Emoji:   "🛑",
		Writer:  cmd.OutOrStdout(),
	})
}
