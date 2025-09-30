package stubs

import (
	"errors"
)

// ConfigManagerStub is a stub implementation of configmanager.ConfigManager[T] interface.
// It provides configurable behavior for testing without external dependencies.
type ConfigManagerStub[T any] struct {
	LoadConfigResult *T
	LoadConfigError  error
	callCount        int
}

// NewConfigManagerStub creates a new ConfigManagerStub with nil result.
func NewConfigManagerStub[T any]() *ConfigManagerStub[T] {
	return &ConfigManagerStub[T]{}
}

// LoadConfig returns the configured result and error.
func (c *ConfigManagerStub[T]) LoadConfig() (*T, error) {
	c.callCount++
	if c.LoadConfigError != nil {
		return nil, c.LoadConfigError
	}
	return c.LoadConfigResult, nil
}

// WithConfig configures the stub to return the specified config.
func (c *ConfigManagerStub[T]) WithConfig(config *T) *ConfigManagerStub[T] {
	c.LoadConfigResult = config
	c.LoadConfigError = nil
	return c
}

// WithError configures the stub to return an error.
func (c *ConfigManagerStub[T]) WithError(err error) *ConfigManagerStub[T] {
	c.LoadConfigError = err
	return c
}

// WithLoadError configures the stub to return a load error.
func (c *ConfigManagerStub[T]) WithLoadError(message string) *ConfigManagerStub[T] {
	c.LoadConfigError = errors.New(message)
	return c
}

// CallCount returns the number of times LoadConfig was called.
func (c *ConfigManagerStub[T]) CallCount() int {
	return c.callCount
}
