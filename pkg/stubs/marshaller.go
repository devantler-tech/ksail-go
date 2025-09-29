package stubs

import (
	"errors"
)

// MarshallerStub is a stub implementation of marshaller.Marshaller[T] interface.
// It provides configurable behavior for testing without external dependencies.
type MarshallerStub[T any] struct {
	MarshalResult      string
	MarshalError       error
	UnmarshalError     error
	UnmarshalStrError  error
	LastMarshalModel   T
	LastUnmarshalData  []byte
	LastUnmarshalStr   string
	callCount          map[string]int
}

// NewMarshallerStub creates a new MarshallerStub with default behavior.
func NewMarshallerStub[T any]() *MarshallerStub[T] {
	return &MarshallerStub[T]{
		MarshalResult: "serialized-data",
		callCount:     make(map[string]int),
	}
}

// Marshal returns the configured result and error, storing the input model.
func (m *MarshallerStub[T]) Marshal(model T) (string, error) {
	m.callCount["Marshal"]++
	m.LastMarshalModel = model
	
	if m.MarshalError != nil {
		return "", m.MarshalError
	}
	return m.MarshalResult, nil
}

// Unmarshal simulates unmarshaling from byte data.
func (m *MarshallerStub[T]) Unmarshal(data []byte, model *T) error {
	m.callCount["Unmarshal"]++
	m.LastUnmarshalData = data
	return m.UnmarshalError
}

// UnmarshalString simulates unmarshaling from string data.
func (m *MarshallerStub[T]) UnmarshalString(data string, model *T) error {
	m.callCount["UnmarshalString"]++
	m.LastUnmarshalStr = data
	return m.UnmarshalStrError
}

// WithMarshalResult configures the stub to return the specified marshaled content.
func (m *MarshallerStub[T]) WithMarshalResult(content string) *MarshallerStub[T] {
	m.MarshalResult = content
	m.MarshalError = nil
	return m
}

// WithMarshalError configures the stub to return an error on Marshal.
func (m *MarshallerStub[T]) WithMarshalError(message string) *MarshallerStub[T] {
	m.MarshalError = errors.New(message)
	return m
}

// WithUnmarshalError configures the stub to return an error on Unmarshal.
func (m *MarshallerStub[T]) WithUnmarshalError(message string) *MarshallerStub[T] {
	m.UnmarshalError = errors.New(message)
	return m
}

// WithUnmarshalStringError configures the stub to return an error on UnmarshalString.
func (m *MarshallerStub[T]) WithUnmarshalStringError(message string) *MarshallerStub[T] {
	m.UnmarshalStrError = errors.New(message)
	return m
}

// CallCount returns the number of times a specific method was called.
func (m *MarshallerStub[T]) CallCount(method string) int {
	return m.callCount[method]
}

// Reset clears all call tracking and results.
func (m *MarshallerStub[T]) Reset() {
	m.callCount = make(map[string]int)
	m.LastUnmarshalData = nil
	m.LastUnmarshalStr = ""
}