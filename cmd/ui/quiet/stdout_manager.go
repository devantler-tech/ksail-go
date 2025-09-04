// Package quiet provides utilities for silencing output.
package quiet

import "os"

// StdoutManager defines an interface for managing stdout operations.
type StdoutManager interface {
	GetStdout() *os.File
	SetStdout(file *os.File)
}

// OSStdoutManager implements StdoutManager using the actual os.Stdout.
type OSStdoutManager struct{}

// GetStdout returns the current os.Stdout.
func (r *OSStdoutManager) GetStdout() *os.File {
	return os.Stdout
}

// SetStdout sets os.Stdout to the provided file.
func (r *OSStdoutManager) SetStdout(file *os.File) {
	os.Stdout = file
}