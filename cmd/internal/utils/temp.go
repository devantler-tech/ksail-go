// Package utils provides common utilities for KSail command creation and handling.
package utils

import (
	"fmt"

	"github.com/spf13/cobra"
)

// HandleConfigLoadRunE provides a shared implementation for commands whose primary
// responsibility is to load the cluster configuration and exit without additional
// side effects. It centralizes timer handling and error wrapping to keep
// individual commands focused on their specific logic.
func HandleConfigLoadRunE(
	cmd *cobra.Command,
	_ []string,
) error {
	// Create command utils
	utils, err := NewCommandUtils(cmd)
	if err != nil {
		return fmt.Errorf("failed to create command utils: %w", err)
	}

	// Start timer
	utils.Timer.Start()

	// Load cluster configuration
	err = utils.ConfigManager.LoadConfig(nil)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	return nil
}
