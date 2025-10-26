// Package configmanager provides centralized configuration management using Viper.
// This file contains the interfaces for configuration management.
package configmanager

import "github.com/devantler-tech/ksail-go/pkg/ui/timer"

// ConfigManager provides configuration management functionality.
//
//go:generate mockery
type ConfigManager[T any] interface {
	// LoadConfig loads the configuration from files and environment variables.
	// Returns the previously loaded config if already loaded.
	// If timer is provided, timing information will be included in the success notification.
	LoadConfig(tmr timer.Timer) error

	// GetConfig returns the currently loaded configuration.
	// If the configuration has not been loaded yet, it returns nil.
	GetConfig() *T
}
