package config

import (
	"github.com/spf13/viper"
)

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