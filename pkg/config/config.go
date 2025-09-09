// Package config provides centralized configuration management using Viper.
package config

import (
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
	v := viper.New()

	// Set configuration file settings
	v.SetConfigName(DefaultConfigFileName)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME")
	v.AddConfigPath("/etc/ksail")

	// Set environment variable settings
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Read configuration file (optional)
	if err := v.ReadInConfig(); err != nil {
		// Configuration file is optional, so we only return error for parsing issues
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Unmarshal into config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// InitializeViper initializes a Viper instance with KSail configuration settings.
// This can be used by commands to bind flags and get configuration values.
func InitializeViper() *viper.Viper {
	v := viper.New()

	// Set configuration file settings
	v.SetConfigName(DefaultConfigFileName)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("$HOME")
	v.AddConfigPath("/etc/ksail")

	// Set environment variable settings
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	// Read configuration file (optional)
	_ = v.ReadInConfig() // Ignore errors for missing config files

	return v
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// Init command defaults
	v.SetDefault("distribution", "Kind")

	// List command defaults
	v.SetDefault("all", false)

	// Cluster defaults
	v.SetDefault("cluster.name", "ksail-default")
	v.SetDefault("cluster.distribution_config", "kind.yaml")
	v.SetDefault("cluster.source_directory", "k8s")
	v.SetDefault("cluster.connection.kubeconfig", "~/.kube/config")
	v.SetDefault("cluster.connection.context", "kind-ksail-default")
	v.SetDefault("cluster.connection.timeout", "5m")
}

// GetConfigFilePath returns the path where the configuration file should be written.
func GetConfigFilePath() string {
	return DefaultConfigFileName + ".yaml"
}