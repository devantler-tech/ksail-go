package cmd

import (
	"fmt"
	"os"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/io/scaffolder"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
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

	selectors := ksailconfigmanager.DefaultClusterFieldSelectors()
	selectors = append(selectors, ksailconfigmanager.StandardSourceDirectoryFieldSelector())
	selectors = append(selectors, ksailconfigmanager.DefaultCNIFieldSelector())
	selectors = append(selectors, ksailconfigmanager.DefaultGitOpsEngineFieldSelector())

	cfgManager := ksailconfigmanager.NewCommandConfigManager(cmd, selectors)

	cmd.Flags().StringP("output", "o", "", "Output directory for the project")
	_ = cfgManager.Viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")
	_ = cfgManager.Viper.BindPFlag("force", cmd.Flags().Lookup("force"))
	cmd.Flags().StringSlice("mirror-registry", []string{}, "Configure mirror registries (e.g., docker.io=http://localhost:5000)")
	_ = cfgManager.Viper.BindPFlag("mirror-registry", cmd.Flags().Lookup("mirror-registry"))

	cmd.RunE = runtime.RunEWithRuntime(
		runtimeContainer,
		runtime.WithTimer(func(cmd *cobra.Command, _ runtime.Injector, tmr timer.Timer) error {
			deps := InitDeps{Timer: tmr}

			return HandleInitRunE(cmd, cfgManager, deps)
		}),
	)

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
	mirrorRegistries := cfgManager.Viper.GetStringSlice("mirror-registry")

	scaffolderInstance := scaffolder.NewScaffolder(
		*cfgManager.GetConfig(),
		cmd.OutOrStdout(),
	)
	scaffolderInstance.MirrorRegistries = mirrorRegistries

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
