// Package config provides centralized configuration management using Viper.
// This file contains the interfaces for configuration management.
package config

import (
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
)

// ConfigManager provides configuration management functionality.
// This is a type alias for backward compatibility.
type ConfigManager = configmanager.ConfigManager[v1alpha1.Cluster]
