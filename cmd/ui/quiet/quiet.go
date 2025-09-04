// Package quiet provides utilities for managing output writers.
package quiet

import (
	"io"
	"os"
)

// GetWriter returns an appropriate writer based on the quiet flag.
// If quiet is true, returns io.Discard to silence output.
// If quiet is false, returns os.Stdout for normal output.
func GetWriter(quiet bool) io.Writer {
	if quiet {
		return io.Discard
	}
	return os.Stdout
}
