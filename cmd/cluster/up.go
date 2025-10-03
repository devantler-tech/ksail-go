package cluster

import (
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultUpTimeout = 5 * time.Minute

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"up",
		"Start the Kubernetes cluster",
		`Start the Kubernetes cluster defined in the project configuration.`,
		HandleUpRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for cluster operations",
			DefaultValue: metav1.Duration{Duration: defaultUpTimeout},
		},
	)
}

// HandleUpRunE handles the up command.
// Exported for testing purposes.
func HandleUpRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Load cluster configuration
	_, err := cmdhelpers.LoadClusterWithErrorHandling(cmd, manager, tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "Cluster created and started successfully (stub implementation)",
		Timer:      tmr,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}
