// Package quiet provides utilities for silencing output.
package quiet

import "os"

// StdoutManager defines an interface for managing stdout operations.
type StdoutManager interface {
	GetStdout() *os.File
	SetStdout(file *os.File)
}

// RealStdoutManager implements StdoutManager using the actual os.Stdout.
type RealStdoutManager struct{}

// GetStdout returns the current os.Stdout.
func (r *RealStdoutManager) GetStdout() *os.File {
	return os.Stdout
}

// SetStdout sets os.Stdout to the provided file.
func (r *RealStdoutManager) SetStdout(file *os.File) {
	os.Stdout = file
}