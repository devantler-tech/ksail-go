package shared

import (
	"fmt"

	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// ConfigLoadDeps provides dependencies required for simple config-loading commands.
type ConfigLoadDeps struct {
	Timer timer.Timer
}

// NewConfigLoaderRunE creates a cobra RunE function that loads the KSail configuration.
func NewConfigLoaderRunE(rt *runtime.Runtime) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout())

		return rt.Invoke(cmd, func(injector do.Injector) error {
			tmr, err := do.Invoke[timer.Timer](injector)
			if err != nil {
				return fmt.Errorf("resolve timer dependency: %w", err)
			}

			deps := ConfigLoadDeps{Timer: tmr}
			return LoadConfig(cmd, cfgManager, deps)
		})
	}
}

// LoadConfig loads the KSail configuration using the provided dependencies.
func LoadConfig(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps ConfigLoadDeps,
) error {
	if deps.Timer != nil {
		deps.Timer.Start()
	}

	if err := cfgManager.LoadConfig(deps.Timer); err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	return nil
}
