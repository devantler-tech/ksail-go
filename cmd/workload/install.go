package workload

import (
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the workload install command.
func NewInstallCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"install",
		"Install Helm charts",
		"Install Helm charts to provision workloads through KSail.",
		HandleInstallRunE,
	)
}

// HandleInstallRunE handles the install command.
func HandleInstallRunE(
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
