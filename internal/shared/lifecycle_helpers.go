package shared

import (
	"context"
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	configmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

var ErrMissingClusterProvisionerDependency = errors.New("missing cluster provisioner dependency")

type LifecycleAction func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error

type LifecycleConfig struct {
	TitleEmoji         string
	TitleContent       string
	ActivityContent    string
	SuccessContent     string
	ErrorMessagePrefix string
	Action             LifecycleAction
}

type LifecycleDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

func NewLifecycleCommandWrapper(
	runtimeContainer *runtime.Runtime,
	cfgManager *ksailconfigmanager.ConfigManager,
	config LifecycleConfig,
) func(*cobra.Command, []string) error {
	return WrapLifecycleHandler(
		runtimeContainer,
		cfgManager,
		func(cmd *cobra.Command, manager *ksailconfigmanager.ConfigManager, deps LifecycleDeps) error {
			return HandleLifecycleRunE(cmd, manager, deps, config)
		},
	)
}

func WrapLifecycleHandler(
	runtimeContainer *runtime.Runtime,
	cfgManager *ksailconfigmanager.ConfigManager,
	handler func(*cobra.Command, *ksailconfigmanager.ConfigManager, LifecycleDeps) error,
) func(*cobra.Command, []string) error {
	return runtime.RunEWithRuntime(
		runtimeContainer,
		runtime.WithTimer(
			func(cmd *cobra.Command, injector runtime.Injector, tmr timer.Timer) error {
				factory, err := runtime.ResolveClusterProvisionerFactory(injector)
				if err != nil {
					return fmt.Errorf("resolve provisioner factory dependency: %w", err)
				}
				deps := LifecycleDeps{Timer: tmr, Factory: factory}
				return handler(cmd, cfgManager, deps)
			},
		),
	)
}

func HandleLifecycleRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps LifecycleDeps,
	config LifecycleConfig,
) error {
	deps.Timer.Start()
	clusterCfg, err := cfgManager.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}
	deps.Timer.NewStage()
	return RunLifecycleWithConfig(cmd, deps, config, clusterCfg)
}

func showLifecycleTitle(cmd *cobra.Command, emoji, content string) {
	cmd.Println()
	notify.WriteMessage(
		notify.Message{
			Type:    notify.TitleType,
			Content: content,
			Emoji:   emoji,
			Writer:  cmd.OutOrStdout(),
		},
	)
}

func RunLifecycleWithConfig(
	cmd *cobra.Command,
	deps LifecycleDeps,
	config LifecycleConfig,
	clusterCfg *v1alpha1.Cluster,
) error {
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
	return runLifecycleWithProvisioner(cmd, deps, config, provisioner, clusterName)
}

func runLifecycleWithProvisioner(
	cmd *cobra.Command,
	deps LifecycleDeps,
	config LifecycleConfig,
	provisioner clusterprovisioner.ClusterProvisioner,
	clusterName string,
) error {
	showLifecycleTitle(cmd, config.TitleEmoji, config.TitleContent)
	notify.WriteMessage(
		notify.Message{
			Type:    notify.ActivityType,
			Content: config.ActivityContent,
			Writer:  cmd.OutOrStdout(),
		},
	)
	if err := config.Action(cmd.Context(), provisioner, clusterName); err != nil {
		return fmt.Errorf("%s: %w", config.ErrorMessagePrefix, err)
	}
	notify.WriteMessage(
		notify.Message{
			Type:    notify.SuccessType,
			Content: config.SuccessContent,
			Writer:  cmd.OutOrStdout(),
		},
	)
	total, stage := deps.Timer.GetTiming()
	timingStr := notify.FormatTiming(total, stage, true)
	// Print timing as part of success context (original code printed after success)
	notify.WriteMessage(
		notify.Message{Type: notify.InfoType, Content: timingStr, Writer: cmd.OutOrStdout()},
	)
	return nil
}
