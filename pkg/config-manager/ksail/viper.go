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
// to other functions. Configuration priority is: defaults < config files < environment variables < flags.
func InitializeViper() *viper.Viper {
	viperInstance := viper.New()

	// Configure file settings first (highest precedence after flags/env)
	configureViperFileSettings(viperInstance)

	// Add standard configuration paths
	configureViperPaths(viperInstance)

	// Setup environment variable handling (higher precedence than config files)
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
// Uses AutomaticEnv() for automatic environment variable binding with proper transformations.
func configureViperEnvironment(viperInstance *viper.Viper) {
	viperInstance.SetEnvPrefix(EnvPrefix)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.AutomaticEnv()

	// Explicitly bind key environment variables for better discoverability
	// AutomaticEnv will handle these automatically, but explicit binding helps with documentation
	bindKnownEnvironmentVariables(viperInstance)
}

// bindKnownEnvironmentVariables explicitly binds known environment variables.
// This is complementary to AutomaticEnv() and helps with IDE autocompletion and documentation.
func bindKnownEnvironmentVariables(viperInstance *viper.Viper) {
	// Map common environment variables to their viper keys
	envMapping := map[string]string{
		"metadata.name":              "KSAIL_METADATA_NAME",
		"spec.distribution":          "KSAIL_SPEC_DISTRIBUTION",
		"spec.sourcedirectory":       "KSAIL_SPEC_SOURCEDIRECTORY",
		"spec.connection.context":    "KSAIL_SPEC_CONNECTION_CONTEXT",
		"spec.connection.kubeconfig": "KSAIL_SPEC_CONNECTION_KUBECONFIG",
		"spec.connection.timeout":    "KSAIL_SPEC_CONNECTION_TIMEOUT",
		"spec.cni":                   "KSAIL_SPEC_CNI",
		"spec.csi":                   "KSAIL_SPEC_CSI",
		"spec.ingresscontroller":     "KSAIL_SPEC_INGRESSCONTROLLER",
		"spec.gatewaycontroller":     "KSAIL_SPEC_GATEWAYCONTROLLER",
		"spec.reconciliationtool":    "KSAIL_SPEC_RECONCILIATIONTOOL",
	}

	for viperKey, envVar := range envMapping {
		_ = viperInstance.BindEnv(viperKey, envVar)
	}
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
