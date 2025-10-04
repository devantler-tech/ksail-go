package workload

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewInstallCommand creates the workload install command.
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
	cmd *cobra.Command,
	_ *configmanager.ConfigManager,
	_ []string,
) error {
	return nil
}
