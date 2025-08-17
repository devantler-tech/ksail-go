// Package quiet provides utilities for silencing output.
package quiet

import "os"

// SilenceStdout runs function with os.Stdout redirected to /dev/null, restoring it afterward.
func SilenceStdout(function func() error) error {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return err
	}
	defer devNull.Close()

	old := os.Stdout
	os.Stdout = devNull

	defer func() { os.Stdout = old }()

	return function()
}
