// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains Viper initialization and configuration utilities.
package ksail

import (
	"os"
	"os/user"
	"path/filepath"
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

// InitializeViper creates a new Viper instance with basic KSail configuration settings.
// This function handles only the essential Viper setup and delegates specific concerns
// to other functions.
func InitializeViper() *viper.Viper {
	viperInstance := viper.New()

	// Delegate configuration setup to specialized functions
	configureViperFileSettings(viperInstance)
	configureViperPaths(viperInstance)
	configureViperEnvironment(viperInstance)

	return viperInstance
}

// configureViperFileSettings sets up file-related configuration for Viper.
func configureViperFileSettings(v *viper.Viper) {
	v.SetConfigName(DefaultConfigFileName)
	v.SetConfigType("yaml")
}

// configureViperPaths adds default configuration search paths to Viper.
func configureViperPaths(viperInstance *viper.Viper) {
	// Get user home directory using os/user instead of $HOME
	usr, err := user.Current()
	if err == nil {
		viperInstance.AddConfigPath(filepath.Join(usr.HomeDir, ".ksail"))
	}

	viperInstance.AddConfigPath("/etc/ksail")
}

// configureViperEnvironment sets up environment variable handling for Viper.
func configureViperEnvironment(v *viper.Viper) {
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
}

// addParentDirectoriesToViperPaths adds parent directories containing ksail.yaml to Viper's search paths.
// This enables directory traversal functionality similar to how Git finds .git directories.
func addParentDirectoriesToViperPaths(viperInstance *viper.Viper) {
	// Get absolute path of current directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		// If we can't get current dir, the default paths should suffice
		return
	}

	// Track which directories we've added to avoid duplicates
	addedPaths := make(map[string]bool)

	// Walk up the directory tree and add each directory to Viper's search paths
	// but only if a ksail.yaml file actually exists in that directory
	for dir := currentDir; ; dir = filepath.Dir(dir) {
		configPath := filepath.Join(dir, "ksail.yaml")

		_, statErr := os.Stat(configPath)
		if statErr == nil {
			// Only add the directory to search path if ksail.yaml exists there
			// and we haven't added it already
			if !addedPaths[dir] {
				viperInstance.AddConfigPath(dir)
				addedPaths[dir] = true
			}
		}

		// Check if we've reached the root directory
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
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
