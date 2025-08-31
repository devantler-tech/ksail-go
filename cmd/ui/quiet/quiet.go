// Package quiet provides utilities for silencing output.
package quiet

import (
	"fmt"
	"os"
)

// FileOpener defines an interface for opening files.
type FileOpener interface {
	Open(name string) (*os.File, error)
}

// DefaultFileOpener implements FileOpener using os.Open.
type DefaultFileOpener struct{}

// Open opens a file using os.Open.
func (d DefaultFileOpener) Open(name string) (*os.File, error) {
	return os.Open(name)
}

// HandleCloseError handles errors from closing files by logging to stderr.
func HandleCloseError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to close os.DevNull: %v\n", err)
	}
}

// SilenceStdout runs function with os.Stdout redirected to /dev/null, restoring it afterward.
func SilenceStdout(function func() error) error {
	return SilenceStdoutWithOpener(DefaultFileOpener{}, function)
}

// SilenceStdoutWithOpener runs function with os.Stdout redirected to /dev/null using the provided opener.
func SilenceStdoutWithOpener(opener FileOpener, function func() error) error {
	devNull, err := opener.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("failed to open os.DevNull: %w", err)
	}

	defer func() {
		HandleCloseError(devNull.Close())
	}()

	old := os.Stdout
	os.Stdout = devNull

	defer func() { os.Stdout = old }()

	return function()
}
