package cmd

import (
	"fmt"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// ConfigLoadDeps captures dependencies required for loading KSail configuration files.
type ConfigLoadDeps struct {
	Timer timer.Timer
}

// NewConfigLoaderRunE returns a cobra RunE that loads the KSail configuration using the runtime container.
func NewConfigLoaderRunE(runtimeContainer *runtime.Runtime) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout())

		return runtimeContainer.Invoke(func(injector runtime.Injector) error {
			tmr, err := do.Invoke[timer.Timer](injector)
			if err != nil {
				return fmt.Errorf("resolve timer dependency: %w", err)
			}

			deps := ConfigLoadDeps{Timer: tmr}

			return LoadConfig(cfgManager, deps)
		})
	}
}

// LoadConfig loads the KSail configuration while tracking timing information.
func LoadConfig(cfgManager *ksailconfigmanager.ConfigManager, deps ConfigLoadDeps) error {
	if deps.Timer != nil {
		deps.Timer.Start()
	}

	_, err := cfgManager.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	return nil
}
