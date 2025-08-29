// Package testutils provides generic marshal failure test utilities.
package testutils

import (
	"errors"

	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
)

// ErrBoom is a common error for testing marshal failures.
var ErrBoom = errors.New("boom")

// MarshalFailer is a generic marshal failer that can be used with any config type.
// It embeds the marshaller interface and overrides only the Marshal method to fail.
type MarshalFailer[T any] struct {
	marshaller.Marshaller[T]
}

// Marshal always returns an error for testing purposes.
func (m MarshalFailer[T]) Marshal(_ T) (string, error) {
	return "", ErrBoom
}
