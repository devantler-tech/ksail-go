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
	// LoadCluster loads the cluster configuration from files and environment variables.
	LoadCluster() (*T, error)

	// GetCluster returns the currently loaded cluster configuration.
	GetCluster() *T

	// GetViper returns the underlying Viper instance for flag binding.
	GetViper() *viper.Viper
}
