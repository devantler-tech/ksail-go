package cluster

import (
	"fmt"

	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/k8sclient"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

// NewStatusCmd creates and returns the status command.
func NewStatusCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Get the status of a cluster",
		Long:         `Get the current status of a Kubernetes cluster.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		return runtimeContainer.Invoke(func(injector runtime.Injector) error {
			tmr, err := do.Invoke[timer.Timer](injector)
			if err != nil {
				return fmt.Errorf("resolve timer dependency: %w", err)
			}

			deps := StatusDeps{
				Timer:                   tmr,
				ClientProvider:          k8sclient.NewDefaultClientProvider(),
				ComponentStatusProvider: k8sclient.NewDefaultComponentStatusProvider(),
			}

			return HandleStatusRunE(cmd, cfgManager, deps)
		})
	}

	return cmd
}

// StatusDeps captures dependencies needed for the status command logic.
type StatusDeps struct {
	Timer                   timer.Timer
	ClientProvider          k8sclient.ClientProvider
	ComponentStatusProvider k8sclient.ComponentStatusProvider
}

// HandleStatusRunE handles the status command.
// Exported for testing purposes.
func HandleStatusRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps StatusDeps,
) error {
	if deps.Timer != nil {
		deps.Timer.Start()
	}

	// Load cluster configuration
	err := cfgManager.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	clusterCfg := cfgManager.GetConfig()

	// Create Kubernetes client
	clientset, err := deps.ClientProvider.CreateClient(
		clusterCfg.Spec.Connection.Kubeconfig,
		clusterCfg.Spec.Connection.Context,
	)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Get component statuses
	statuses, err := deps.ComponentStatusProvider.GetComponentStatuses(
		cmd.Context(),
		clientset,
	)
	if err != nil {
		return fmt.Errorf("failed to get component statuses: %w", err)
	}

	// Display component statuses
	displayComponentStatuses(cmd, statuses)

	// Display timing if timer exists
	if deps.Timer != nil {
		total, stage := deps.Timer.GetTiming()
		timingStr := notify.FormatTiming(total, stage, false)
		notify.WriteMessage(notify.Message{
			Type:    notify.SuccessType,
			Content: "cluster status retrieved " + timingStr,
			Writer:  cmd.OutOrStdout(),
		})
	}

	return nil
}

func displayComponentStatuses(cmd *cobra.Command, statuses []corev1.ComponentStatus) {
	if len(statuses) == 0 {
		notify.WriteMessage(notify.Message{
			Type:    notify.ActivityType,
			Content: "no component statuses found",
			Writer:  cmd.OutOrStdout(),
		})

		return
	}

	// Display header
	fmt.Fprintf(cmd.OutOrStdout(), "NAME                 STATUS    MESSAGE\n")

	// Display each component
	for _, status := range statuses {
		statusStr := "Unknown"
		message := ""

		for _, condition := range status.Conditions {
			if condition.Type == corev1.ComponentHealthy {
				if condition.Status == corev1.ConditionTrue {
					statusStr = "Healthy"
				} else {
					statusStr = "Unhealthy"
				}

				message = condition.Message

				break
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-9s %s\n", status.Name, statusStr, message)
	}
}
