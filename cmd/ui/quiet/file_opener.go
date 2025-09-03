// Package quiet provides utilities for silencing output.
package quiet

import "os"

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