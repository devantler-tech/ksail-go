// Package config provides centralized configuration management using Viper.
// This file contains the interfaces for configuration management.
package config

import (
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/viper"
)

// ConfigManager provides configuration management functionality.
//
//go:generate mockery
type ConfigManager interface {
	// LoadCluster loads the cluster configuration from files and environment variables.
	LoadCluster() (*v1alpha1.Cluster, error)

	// GetCluster returns the currently loaded cluster configuration.
	GetCluster() *v1alpha1.Cluster

	// GetViper returns the underlying Viper instance for flag binding.
	GetViper() *viper.Viper
}
