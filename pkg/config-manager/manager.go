// Package configmanager provides centralized configuration management using Viper.
// This file contains the interfaces for configuration management.
package configmanager

import (
	"github.com/spf13/viper"
)

// ConfigManager provides configuration management functionality.
//
//go:generate mockery
type ConfigManager[T any] interface {
	// LoadConfig loads the configuration from files and environment variables.
	LoadConfig() (*T, error)

	// GetConfig returns the currently loaded configuration.
	GetConfig() *T

	// GetViper returns the underlying Viper instance for flag binding.
	GetViper() *viper.Viper
}
