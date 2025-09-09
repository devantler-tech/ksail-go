// Package config provides centralized configuration management using Viper.
package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const (
	// DefaultConfigFileName is the default configuration file name (without extension).
	DefaultConfigFileName = "ksail"
	// EnvPrefix is the prefix for environment variables.
	EnvPrefix = "KSAIL"
)

// Config holds all configuration values for KSail.
type Config struct {
	// Init command configuration
	Distribution string `mapstructure:"distribution"`

	// List command configuration
	All bool `mapstructure:"all"`

	// Cluster configuration defaults
	Cluster ClusterConfig `mapstructure:"cluster"`
}

// ClusterConfig holds default cluster configuration values.
type ClusterConfig struct {
	Name               string `mapstructure:"name"`
	DistributionConfig string `mapstructure:"distribution_config"`
	SourceDirectory    string `mapstructure:"source_directory"`
	Connection         ConnectionConfig `mapstructure:"connection"`
}

// ConnectionConfig holds default connection configuration values.
type ConnectionConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
	Context    string `mapstructure:"context"`
	Timeout    string `mapstructure:"timeout"`
}

// LoadConfig loads configuration from files, environment variables, and sets defaults.
func LoadConfig() (*Config, error) {
	viperInstance := viper.New()

	// Set configuration file settings
	viperInstance.SetConfigName(DefaultConfigFileName)
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(".")
	viperInstance.AddConfigPath("$HOME")
	viperInstance.AddConfigPath("/etc/ksail")

	// Set environment variable settings
	viperInstance.SetEnvPrefix(EnvPrefix)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.AutomaticEnv()

	// Set defaults
	setDefaults(viperInstance)

	// Read configuration file (optional)
	err := viperInstance.ReadInConfig()
	if err != nil {
		// Configuration file is optional, so we only return error for parsing issues
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Unmarshal into config struct
	var config Config
	err = viperInstance.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// InitializeViper initializes a Viper instance with KSail configuration settings.
// This can be used by commands to bind flags and get configuration values.
func InitializeViper() *viper.Viper {
	viperInstance := viper.New()

	// Set configuration file settings
	viperInstance.SetConfigName(DefaultConfigFileName)
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(".")
	viperInstance.AddConfigPath("$HOME")
	viperInstance.AddConfigPath("/etc/ksail")

	// Set environment variable settings
	viperInstance.SetEnvPrefix(EnvPrefix)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viperInstance.AutomaticEnv()

	// Set defaults
	setDefaults(viperInstance)

	// Read configuration file (optional)
	_ = viperInstance.ReadInConfig() // Ignore errors for missing config files

	return viperInstance
}

// setDefaults sets default configuration values.
func setDefaults(viperInstance *viper.Viper) {
	// Init command defaults
	viperInstance.SetDefault("distribution", "Kind")

	// List command defaults
	viperInstance.SetDefault("all", false)

	// Cluster defaults
	viperInstance.SetDefault("cluster.name", "ksail-default")
	viperInstance.SetDefault("cluster.distribution_config", "kind.yaml")
	viperInstance.SetDefault("cluster.source_directory", "k8s")
	viperInstance.SetDefault("cluster.connection.kubeconfig", "~/.kube/config")
	viperInstance.SetDefault("cluster.connection.context", "kind-ksail-default")
	viperInstance.SetDefault("cluster.connection.timeout", "5m")
}

// GetConfigFilePath returns the path where the configuration file should be written.
func GetConfigFilePath() string {
	return DefaultConfigFileName + ".yaml"
}