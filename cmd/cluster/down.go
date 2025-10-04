package cluster

import (
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		HandleDownRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardContextFieldSelector(),
	)
}

// HandleDownRunE handles the down command.
func HandleDownRunE(
	_ *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	tmr := timer.New()
	tmr.Start()

	_, err := manager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	return nil
}
