package stubs

import (
	"errors"
	"os"
	"path/filepath"

	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
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

	content := g.GenerateResult

	// If this is a yamlgenerator.Options and has an Output path, write the file
	if outputOpts, ok := any(opts).(yamlgenerator.Options); ok && outputOpts.Output != "" {
		// Write the content to the file
		err := writeStubFile(content, outputOpts.Output, outputOpts.Force)
		if err != nil {
			return "", err
		}
	}

	return content, nil
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

// writeStubFile writes content to a file, creating directories as needed.
func writeStubFile(content, outputPath string, force bool) error {
	// Check if file exists and force flag
	if _, err := os.Stat(outputPath); err == nil && !force {
		return errors.New("file already exists and force is false")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Write the file
	return os.WriteFile(outputPath, []byte(content), 0o644)
}
