// Package quiet provides utilities for silencing output.
package quiet

import "os"

// FileOpener defines an interface for opening files.
type FileOpener interface {
	Open(name string) (*os.File, error)
}