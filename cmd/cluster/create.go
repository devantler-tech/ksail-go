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

// NewCreateCmd wires the cluster create command using the shared runtime container.
func NewCreateCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "Create a cluster",
		Long:         `Create a Kubernetes cluster as defined by configuration.`,
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

				deps := CreateDeps{
					Timer:   tmr,
					Factory: factory,
				}

				return HandleCreateRunE(cmd, cfgManager, deps)
			},
		),
	)

	return cmd
}

// CreateDeps contains the dependencies required to handle the create command.
type CreateDeps struct {
	Timer   timer.Timer
	Factory clusterprovisioner.Factory
}

var errMissingClusterProvisioner = errors.New("missing cluster provisioner dependency")

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
		return errMissingClusterProvisioner
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

	err = provisioner.Create(cmd.Context(), clusterName)
	if err != nil {
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
