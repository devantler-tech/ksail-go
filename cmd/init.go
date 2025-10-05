package cmd

import (
	"fmt"
	"os"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	// Create the command using the helper
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Initialize a new project",
		Long:         "Initialize a new project in the specified directory (or current directory if none specified).",
		SilenceUsage: true,
	}

	// Create command utils
	utils, _ := utils.NewCommandUtils(cmd,
		configmanager.DefaultDistributionFieldSelector(),
		configmanager.DefaultDistributionConfigFieldSelector(),
		configmanager.StandardSourceDirectoryFieldSelector())

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return HandleInitRunE(cmd, utils, args)
	}

	// Bind CLI only flags
	_ = utils.ConfigManager.Viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	cmd.Flags().StringP("output", "o", "", "Output directory for the project")
	_ = utils.ConfigManager.Viper.BindPFlag("force", cmd.Flags().Lookup("force"))
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")

	return cmd
}

// HandleInitRunE handles the init command with an optional output path.
// If outputPath is empty, uses the current working directory.
// The variadic outputPath parameter is for testing purposes only.
// Exported for testing purposes.
func HandleInitRunE(
	cmd *cobra.Command,
	utils *utils.CommandUtils,
	_ []string,
) error {
	// Start timing
	utils.Timer.Start()

	// Load the configuration
	err := utils.ConfigManager.LoadConfig(utils.Timer)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Get output path
	var targetPath string
	flagOutputPath := utils.ConfigManager.Viper.GetString("output")
	if flagOutputPath != "" {
		targetPath = flagOutputPath
	} else {
		targetPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Get the force flag value
	force := utils.ConfigManager.Viper.GetBool("force")

	// Create scaffolder and generate project files
	scaffolderInstance := scaffolder.NewScaffolder(
		*utils.ConfigManager.GetConfig(),
		cmd.OutOrStdout(),
	)

	cmd.Println()

	// Mark new stage for scaffolding
	utils.Timer.NewStage()

	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Initialize project......",
		Emoji:   "📂",
		Writer:  cmd.OutOrStdout(),
	})

	// Generate files individually to provide immediate feedback
	err = scaffolderInstance.Scaffold(targetPath, force)
	if err != nil {
		return fmt.Errorf("failed to scaffold project files: %w", err)
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "initialized project",
		Timer:   utils.Timer,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}
