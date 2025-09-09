// Package config provides centralized configuration management using Viper.
package config

import (
	"strings"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/internal/utils/k8s"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultConfigFileName is the default configuration file name (without extension).
	DefaultConfigFileName = "ksail"
	// EnvPrefix is the prefix for environment variables.
	EnvPrefix = "KSAIL"
)

// Manager provides configuration management functionality using the v1alpha1.Cluster structure.
type Manager struct {
	viper   *viper.Viper
	cluster *v1alpha1.Cluster
}

// NewManager creates a new configuration manager.
func NewManager() *Manager {
	return &Manager{
		viper:   initializeViper(),
		cluster: nil,
	}
}

// LoadCluster loads the cluster configuration from files and environment variables.
func (m *Manager) LoadCluster() (*v1alpha1.Cluster, error) {
	// Create a cluster WITHOUT defaults first
	cluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.Kind,
			APIVersion: v1alpha1.APIVersion,
		},
		Metadata: k8s.NewEmptyObjectMeta(),
		Spec:     v1alpha1.Spec{},
	}

	// Apply configuration from the config source FIRST
	cluster.SetDefaultsFromConfigSource(m.viper)

	// Then apply any missing defaults
	cluster.SetDefaults()

	// Store the loaded cluster
	m.cluster = cluster

	return cluster, nil
}

// GetCluster returns the currently loaded cluster configuration.
func (m *Manager) GetCluster() *v1alpha1.Cluster {
	if m.cluster == nil {
		// Return a default cluster if none is loaded
		return v1alpha1.NewCluster()
	}
	return m.cluster
}

// GetViper returns the underlying Viper instance for flag binding.
func (m *Manager) GetViper() *viper.Viper {
	return m.viper
}

// GetString gets a configuration value as string.
func (m *Manager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetBool gets a configuration value as bool.
func (m *Manager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// initializeViper initializes a Viper instance with KSail configuration settings.
func initializeViper() *viper.Viper {
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

	// Read configuration file (optional)
	_ = viperInstance.ReadInConfig() // Ignore errors for missing config files

	return viperInstance
}

// GetConfigFilePath returns the path where the configuration file should be written.
func GetConfigFilePath() string {
	return DefaultConfigFileName + ".yaml"
}

// LoadConfig loads configuration and returns both CLI config and cluster config.
// Deprecated: Use Manager instead for better separation of concerns.
func LoadConfig() (*Config, error) {
	manager := NewManager()
	cluster, err := manager.LoadCluster()
	if err != nil {
		return nil, err
	}

	return &Config{
		Distribution: manager.GetString("distribution"),
		All:          manager.GetBool("all"),
		Cluster: ClusterConfig{
			Name:               cluster.Metadata.Name,
			DistributionConfig: cluster.Spec.DistributionConfig,
			SourceDirectory:    cluster.Spec.SourceDirectory,
			Connection: ConnectionConfig{
				Kubeconfig: cluster.Spec.Connection.Kubeconfig,
				Context:    cluster.Spec.Connection.Context,
				Timeout:    "5m", // Use original format for backward compatibility
			},
		},
	}, nil
}

// InitializeViper initializes a Viper instance with KSail configuration settings.
// Deprecated: Use Manager.GetViper() instead.
func InitializeViper() *viper.Viper {
	return initializeViper()
}

// Config holds all configuration values for KSail.
// Deprecated: Use Manager and v1alpha1.Cluster directly instead.
type Config struct {
	// CLI-specific configuration
	Distribution string `mapstructure:"distribution"`
	All          bool   `mapstructure:"all"`

	// Cluster configuration using a flat structure for backward compatibility
	Cluster ClusterConfig `mapstructure:"cluster"`
}

// ClusterConfig holds default cluster configuration values for backward compatibility.
// Deprecated: Use v1alpha1.Cluster instead.
type ClusterConfig struct {
	Name               string           `mapstructure:"name"`
	DistributionConfig string           `mapstructure:"distribution_config"`
	SourceDirectory    string           `mapstructure:"source_directory"`
	Connection         ConnectionConfig `mapstructure:"connection"`
}

// ConnectionConfig holds default connection configuration values for backward compatibility.
// Deprecated: Use v1alpha1.Connection instead.
type ConnectionConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
	Context    string `mapstructure:"context"`
	Timeout    string `mapstructure:"timeout"`
}