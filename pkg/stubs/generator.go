package stubs

import (
	"errors"
)

// GeneratorStub is a stub implementation of generator.Generator[T, Options] interface.
// It provides configurable behavior for testing without external dependencies.
type GeneratorStub[T any, Options any] struct {
	GenerateResult string
	GenerateError  error
	LastModel      T
	LastOptions    Options
	callCount      int
}

// NewGeneratorStub creates a new GeneratorStub with default behavior.
func NewGeneratorStub[T any, Options any]() *GeneratorStub[T, Options] {
	return &GeneratorStub[T, Options]{
		GenerateResult: "# Generated content\ntest: value\n",
	}
}

// Generate returns the configured result and error, storing the input parameters.
func (g *GeneratorStub[T, Options]) Generate(model T, opts Options) (string, error) {
	g.callCount++
	g.LastModel = model
	g.LastOptions = opts
	
	if g.GenerateError != nil {
		return "", g.GenerateError
	}
	return g.GenerateResult, nil
}

// WithResult configures the stub to return the specified content.
func (g *GeneratorStub[T, Options]) WithResult(content string) *GeneratorStub[T, Options] {
	g.GenerateResult = content
	g.GenerateError = nil
	return g
}

// WithError configures the stub to return an error.
func (g *GeneratorStub[T, Options]) WithError(err error) *GeneratorStub[T, Options] {
	g.GenerateError = err
	return g
}

// WithGenerationError configures the stub to return a generation error.
func (g *GeneratorStub[T, Options]) WithGenerationError(message string) *GeneratorStub[T, Options] {
	g.GenerateError = errors.New(message)
	return g
}

// CallCount returns the number of times Generate was called.
func (g *GeneratorStub[T, Options]) CallCount() int {
	return g.callCount
}