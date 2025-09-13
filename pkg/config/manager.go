// Package config provides centralized configuration management using Viper.
// This file contains the main configuration Manager for handling cluster configuration.
package config

import (
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
)

// Manager provides configuration management functionality using the v1alpha1.Cluster structure.
// This is a type alias for backward compatibility.
type Manager = ksail.Manager

// NewManager creates a new configuration manager.
// Field selectors are optional - if none provided, manager works for commands without configuration needs.
func NewManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *Manager {
	// Convert old field selectors to new field selectors
	newFieldSelectors := make([]ksail.FieldSelector[v1alpha1.Cluster], len(fieldSelectors))
	for i, fs := range fieldSelectors {
		newFieldSelectors[i] = ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     fs.Selector,
			Description:  fs.Description,
			DefaultValue: fs.DefaultValue,
		}
	}
	return ksail.NewManager(newFieldSelectors...)
}
