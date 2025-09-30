// Package stubconfig provides utilities for managing stub configuration in CLI commands.
// This package enables switching between real and stub implementations based on CLI flags.
package stubconfig

import (
	"github.com/spf13/cobra"
)

// IsStubMode checks if the current command has the --stub flag enabled.
// This function traverses up the command hierarchy to find the global --stub flag.
func IsStubMode(cmd *cobra.Command) bool {
	// Start with the current command and traverse up to find the root
	current := cmd
	for current != nil {
		if stubFlag := current.Flags().Lookup("stub"); stubFlag != nil {
			if stubFlag.Value.String() == "true" {
				return true
			}
		}
		if persistentStubFlag := current.PersistentFlags().Lookup("stub"); persistentStubFlag != nil {
			if persistentStubFlag.Value.String() == "true" {
				return true
			}
		}
		current = current.Parent()
	}
	return false
}