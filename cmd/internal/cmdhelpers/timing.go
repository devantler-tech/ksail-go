package cmdhelpers

import (
	"fmt"

	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// ExecuteTimedClusterCommand is a helper that handles timing for simple cluster commands
// that follow the pattern: load cluster → execute → report success with timing.
func ExecuteTimedClusterCommand(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	successMessage string,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Load cluster and execute
	_, err := LoadClusterWithErrorHandling(cmd, manager, tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Display success with timing
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: successMessage,
		Timer:   tmr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}
