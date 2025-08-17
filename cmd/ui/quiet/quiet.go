// Package quiet provides utilities for silencing output.
package quiet

import (
	"fmt"
	"os"
)

// SilenceStdout runs function with os.Stdout redirected to /dev/null, restoring it afterward.
func SilenceStdout(function func() error) error {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("failed to open os.DevNull: %w", err)
	}

	defer func() {
		err := devNull.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to close os.DevNull: %v\n", err)
		}
	}()

	old := os.Stdout
	os.Stdout = devNull

	defer func() { os.Stdout = old }()

	return function()
}
