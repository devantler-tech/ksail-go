// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains Viper initialization and configuration constants.
package ksail

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigFileName is the default configuration file name (without extension).
	DefaultConfigFileName = "ksail"
	// EnvPrefix is the prefix for environment variables.
	EnvPrefix = "KSAIL"
	// SuggestionsMinimumDistance represents the minimum levenshtein distance for command suggestions.
	SuggestionsMinimumDistance = 2
)

// InitializeViper initializes a Viper instance with KSail configuration settings.
func InitializeViper() *viper.Viper {
	viperInstance := viper.New()

	// Set configuration file settings
	viperInstance.SetConfigName(DefaultConfigFileName)
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(".")
	viperInstance.AddConfigPath("$HOME/.config/ksail")
	viperInstance.AddConfigPath("/etc/ksail")

	// Set environment variable settings
	viperInstance.SetEnvPrefix(EnvPrefix)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.AutomaticEnv()

	// Read configuration file (optional)
	_ = viperInstance.ReadInConfig() // Ignore errors for missing config files

	return viperInstance
}

// bindEnvironmentVariables binds environment variables to their corresponding viper keys.
func bindEnvironmentVariables(viperInstance *viper.Viper) {
	// Map common environment variables to their viper keys
	envMapping := map[string]string{
		"METADATA_NAME":              "metadata.name",
		"SPEC_DISTRIBUTION":          "spec.distribution",
		"SPEC_SOURCEDIRECTORY":       "spec.sourcedirectory",
		"SPEC_CONNECTION_CONTEXT":    "spec.connection.context",
		"SPEC_CONNECTION_KUBECONFIG": "spec.connection.kubeconfig",
		"SPEC_CONNECTION_TIMEOUT":    "spec.connection.timeout",
		"SPEC_CNI":                   "spec.cni",
		"SPEC_CSI":                   "spec.csi",
		"SPEC_INGRESSCONTROLLER":     "spec.ingresscontroller",
		"SPEC_GATEWAYCONTROLLER":     "spec.gatewaycontroller",
		"SPEC_RECONCILIATIONTOOL":    "spec.reconciliationtool",
	}

	for envKey, viperKey := range envMapping {
		_ = viperInstance.BindEnv(viperKey, "KSAIL_"+envKey)
	}
}
