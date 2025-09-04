// Package quiet provides utilities for silencing output.
package quiet

import (
	"fmt"
	"os"
)

// handleCloseError handles errors from closing files by logging to stderr.
func handleCloseError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to close os.DevNull: %v\n", err)
	}
}

// SilenceStdout runs function with stdout redirected to /dev/null using the provided opener and stdout manager.
func SilenceStdout(opener FileOpener, stdoutManager StdoutManager, function func() error) error {
	devNull, err := opener.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("failed to open os.DevNull: %w", err)
	}

	defer func() {
		handleCloseError(devNull.Close())
	}()

	old := stdoutManager.GetStdout()
	stdoutManager.SetStdout(devNull)

	defer func() { stdoutManager.SetStdout(old) }()

	return function()
}
