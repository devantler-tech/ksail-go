// Package testutils provides generic marshal failure test utilities.
package testutils

import "errors"

// ErrBoom is a common error for testing marshal failures.
var ErrBoom = errors.New("boom")

// MarshallerInterface represents a generic marshaller interface for testing.
type MarshallerInterface[T any] interface {
	Marshal(config T) (string, error)
	Unmarshal(data []byte, model *T) error
	UnmarshalString(data string, model *T) error
}

// MarshalFailer is a generic marshal failer that can be used with any config type.
// It embeds the marshaller interface and overrides only the Marshal method to fail.
type MarshalFailer[T any] struct {
	MarshallerInterface[T]
}

// Marshal always returns an error for testing purposes.
func (m MarshalFailer[T]) Marshal(_ T) (string, error) {
	return "", ErrBoom
}

// Unmarshal placeholder implementation (not used in tests).
func (m MarshalFailer[T]) Unmarshal(_ []byte, _ *T) error {
	return nil
}

// UnmarshalString placeholder implementation (not used in tests).
func (m MarshalFailer[T]) UnmarshalString(_ string, _ *T) error {
	return nil
}