package shared

import (
	"context"
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

// ErrMissingClusterProvisionerDependency is returned when the cluster provisioner is nil.
var ErrMissingClusterProvisionerDependency = errors.New("missing cluster provisioner dependency")

// LifecycleAction defines a function that performs an action on a cluster provisioner.
type LifecycleAction func(
	ctx context.Context,
	provisioner clusterprovisioner.ClusterProvisioner,
	clusterName string,
) error

// LifecycleConfig contains the configuration for a lifecycle operation.
type LifecycleConfig struct {
	TitleEmoji         string
	TitleContent       string
	ActivityContent    string
	SuccessContent     string
	ErrorMessagePrefix string
	Action             LifecycleAction
}

// LifecycleDeps contains the dependencies required to handle lifecycle commands.
type LifecycleDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

// NewLifecycleCommandWrapper creates a command wrapper for lifecycle operations.
func NewLifecycleCommandWrapper(
	runtimeContainer *runtime.Runtime,
	cfgManager *ksailconfigmanager.ConfigManager,
	config LifecycleConfig,
) func(*cobra.Command, []string) error {
	return runtime.RunEWithRuntime(
		runtimeContainer,
		runtime.WithTimer(
			func(cmd *cobra.Command, injector runtime.Injector, tmr timer.Timer) error {
				factory, err := runtime.ResolveClusterProvisionerFactory(injector)
				if err != nil {
					return fmt.Errorf("resolve provisioner factory dependency: %w", err)
				}

				deps := LifecycleDeps{
					Timer:   tmr,
					Factory: factory,
				}

				return HandleLifecycleRunE(cmd, cfgManager, deps, config)
			},
		),
	)
}

// HandleLifecycleRunE executes a cluster lifecycle workflow.
func HandleLifecycleRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps LifecycleDeps,
	config LifecycleConfig,
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
		return ErrMissingClusterProvisionerDependency
	}

	clusterName, err := configmanager.GetClusterName(distributionConfig)
	if err != nil {
		return fmt.Errorf("failed to get cluster name from config: %w", err)
	}

	showLifecycleTitle(cmd, config.TitleEmoji, config.TitleContent)

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: config.ActivityContent,
		Writer:  cmd.OutOrStdout(),
	})

	err = config.Action(cmd.Context(), provisioner, clusterName)
	if err != nil {
		return fmt.Errorf("%s: %w", config.ErrorMessagePrefix, err)
	}

	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    config.SuccessContent,
		Timer:      deps.Timer,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// showLifecycleTitle displays the lifecycle stage title.
func showLifecycleTitle(cmd *cobra.Command, emoji, content string) {
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: content,
		Emoji:   emoji,
		Writer:  cmd.OutOrStdout(),
	})
}
