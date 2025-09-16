// Package configmanager provides centralized configuration management using Viper.
// This file contains the interfaces for configuration management.
package configmanager

// ConfigManager provides configuration management functionality.
//
//go:generate mockery
type ConfigManager[T any] interface {
	// LoadConfig loads the configuration from files and environment variables.
	// Returns the previously loaded config if already loaded.
	LoadConfig() (*T, error)
}
