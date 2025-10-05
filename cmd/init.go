package cmd

import (
	"fmt"
	"os"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardDistributionConfigFieldSelector(),
		configmanager.StandardSourceDirectoryFieldSelector(),
	}

	// Create the command using the helper
	cmd := helpers.NewCobraCommand(
		"init",
		"Initialize a new project",
		"Initialize a new project.",
		HandleInitRunE,
		fieldSelectors...,
	)

	// Add the --output flag for specifying output directory
	cmd.Flags().StringP("output", "o", "", "Output directory for the project")

	// Add the --force flag for overwriting existing files
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing files")

	return cmd
}

// HandleInitRunE handles the init command with an optional output path.
// If outputPath is empty, uses the current working directory.
// The variadic outputPath parameter is for testing purposes only.
// Exported for testing purposes.
func HandleInitRunE(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	_ []string,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Bind the --output and --force flags
	_ = configManager.Viper.BindPFlag("output", cmd.Flags().Lookup("output"))
	_ = configManager.Viper.BindPFlag("force", cmd.Flags().Lookup("force"))

	cluster, err := configManager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Determine target path - prioritize test parameter over flag
	var targetPath string
	// Get output path from flag
	flagOutputPath := configManager.Viper.GetString("output")
	if flagOutputPath != "" {
		targetPath = flagOutputPath
	} else {
		targetPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Get the force flag value
	force := configManager.Viper.GetBool("force")

	// Create scaffolder and generate project files
	scaffolderInstance := scaffolder.NewScaffolder(*cluster, cmd.OutOrStdout())

	cmd.Println()

	// Mark new stage for scaffolding
	tmr.NewStage()

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
		Timer:   tmr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}
