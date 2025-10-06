package cmd

import (
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Initialize a new project",
		Long:         "Initialize a new project in the specified directory (or current directory if none specified).",
		SilenceUsage: true,
	}

	selectors := []ksailconfigmanager.FieldSelector[v1alpha1.Cluster]{
		ksailconfigmanager.DefaultDistributionFieldSelector(),
		ksailconfigmanager.DefaultDistributionConfigFieldSelector(),
		ksailconfigmanager.StandardSourceDirectoryFieldSelector(),
	}

	cfgManager := ksailconfigmanager.NewConfigManager(cmd.OutOrStdout(), selectors...)
	cfgManager.AddFlagsFromFields(cmd)

	cmd.Flags().StringP("output", "o", "", "Output directory for the project")
	_ = cfgManager.Viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	_ = cfgManager.Viper.BindPFlag("force", cmd.Flags().Lookup("force"))

	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		return runtimeContainer.Invoke(func(injector runtime.Injector) error {
			tmr, err := do.Invoke[timer.Timer](injector)
			if err != nil {
				return fmt.Errorf("resolve timer dependency: %w", err)
			}

			deps := InitDeps{Timer: tmr}

			return HandleInitRunE(cmd, cfgManager, deps)
		})
	}

	return cmd
}

// InitDeps captures dependencies required for the init command.
type InitDeps struct {
	Timer timer.Timer
}

// HandleInitRunE handles the init command.
func HandleInitRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps InitDeps,
) error {
	if deps.Timer != nil {
		deps.Timer.Start()
	}

	err := cfgManager.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	var targetPath string

	flagOutputPath := cfgManager.Viper.GetString("output")
	if flagOutputPath != "" {
		targetPath = flagOutputPath
	} else {
		targetPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	force := cfgManager.Viper.GetBool("force")

	scaffolderInstance := scaffolder.NewScaffolder(
		*cfgManager.GetConfig(),
		cmd.OutOrStdout(),
	)

	cmd.Println()

	if deps.Timer != nil {
		deps.Timer.NewStage()
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Initialize project......",
		Emoji:   "📂",
		Writer:  cmd.OutOrStdout(),
	})

	err = scaffolderInstance.Scaffold(targetPath, force)
	if err != nil {
		return fmt.Errorf("failed to scaffold project files: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "initialized project",
		Timer:   deps.Timer,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}
