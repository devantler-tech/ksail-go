package cluster

import (
	"context"
	"fmt"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// LifecycleOperation defines a function that performs an operation on a cluster provisioner.
type LifecycleOperation func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error

// LifecycleConfig contains configuration for lifecycle command execution.
type LifecycleConfig struct {
	TitleEmoji      string
	TitleContent    string
	ActivityContent string
	SuccessContent  string
	ErrorPrefix     string
}

// LifecycleDeps contains the dependencies required for lifecycle commands.
type LifecycleDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

// LifecycleHandler is a function that executes lifecycle command logic.
type LifecycleHandler func(cmd *cobra.Command, cfgManager *ksailconfigmanager.ConfigManager, deps LifecycleDeps) error

// NewLifecycleCmd creates a lifecycle command with standard wiring.
func NewLifecycleCmd(
	runtimeContainer *runtime.Runtime,
	use, short, long string,
	handler LifecycleHandler,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:          use,
		Short:        short,
		Long:         long,
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

				deps := LifecycleDeps{
					Timer:   tmr,
					Factory: factory,
				}

				return handler(cmd, cfgManager, deps)
			},
		),
	)

	return cmd
}

// ExecuteLifecycleCommand is a shared helper that executes common lifecycle workflow.
func ExecuteLifecycleCommand(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps LifecycleDeps,
	config LifecycleConfig,
	operation LifecycleOperation,
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
		return fmt.Errorf("failed to get cluster name: %w", err)
	}

	showLifecycleTitle(cmd, config.TitleEmoji, config.TitleContent)

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: config.ActivityContent,
		Writer:  cmd.OutOrStdout(),
	})

	err = operation(cmd.Context(), provisioner, clusterName)
	if err != nil {
		return fmt.Errorf("%s: %w", config.ErrorPrefix, err)
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
