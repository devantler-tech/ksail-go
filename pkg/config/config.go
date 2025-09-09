// Package config provides centralized configuration management using Viper.
package config

import (
	"strings"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/internal/utils/k8s"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultConfigFileName is the default configuration file name (without extension).
	DefaultConfigFileName = "ksail"
	// EnvPrefix is the prefix for environment variables.
	EnvPrefix = "KSAIL"
	// SuggestionsMinimumDistance is the minimum edit distance to suggest a command.
	SuggestionsMinimumDistance = 2
)


// NewCobraCommand creates a cobra.Command with automatic configuration binding.
// This is the only constructor provided for initializing CobraCommands with configuration field paths.
// The binding automatically handles CLI flags (priority 1), environment variables (priority 2), 
// and configuration defaults (priority 3).
func NewCobraCommand(
	use, short, long string,
	runE func(*cobra.Command, *Manager, []string) error,
	configPaths []string,
) *cobra.Command {
	manager := NewManager()
	
	// Create the base command
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(cmd, manager, args)
		},
		SuggestionsMinimumDistance: SuggestionsMinimumDistance,
	}
	
	// Bind flags based on configuration paths
	if len(configPaths) > 0 {
		bindConfigurationPaths(cmd, manager, configPaths)
	}
	
	return cmd
}

// bindConfigurationPaths binds CLI flags for the specified configuration paths.
func bindConfigurationPaths(cmd *cobra.Command, manager *Manager, configPaths []string) {
	for _, path := range configPaths {
		flagName, description := getConfigPathInfo(path)
		if flagName == "" {
			continue
		}
		
		// Handle special CLI-only flags
		if flagName == "all" {
			cmd.Flags().Bool("all", false, description)
			_ = manager.viper.BindPFlag("all", cmd.Flags().Lookup("all"))
			continue
		}
		
		// Add string flag for configuration fields
		cmd.Flags().String(flagName, "", description)
		_ = manager.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
	}
}

// getConfigPathInfo maps configuration paths to flag names and descriptions.
func getConfigPathInfo(configPath string) (flagName, description string) {
	switch configPath {
	case "spec.distribution":
		return "distribution", "Configure cluster distribution (Kind, K3d, EKS, Tind)"
	case "spec.distributionConfig":
		return "distribution-config", "Configure distribution config file path"
	case "spec.sourceDirectory":
		return "source-directory", "Configure source directory for Kubernetes manifests"
	case "spec.connection.kubeconfig":
		return "connection-kubeconfig", "Configure kubeconfig file path"
	case "spec.connection.context":
		return "connection-context", "Configure kubectl context"
	case "spec.connection.timeout":
		return "connection-timeout", "Configure connection timeout duration"
	case "spec.cni":
		return "c-n-i", "Configure CNI (Container Network Interface)"
	case "spec.csi":
		return "c-s-i", "Configure CSI (Container Storage Interface)"
	case "spec.ingressController":
		return "ingress-controller", "Configure ingress controller"
	case "spec.gatewayController":
		return "gateway-controller", "Configure gateway controller"
	case "spec.reconciliationTool":
		return "reconciliation-tool", "Configure reconciliation tool (Kubectl, Flux, ArgoCD)"
	case "metadata.name":
		return "name", "Configure cluster name"
	case "all":
		return "all", "List all clusters including stopped ones"
	default:
		return "", ""
	}
}



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

	// Apply all configuration and defaults through the config manager
	m.setClusterFromConfig(cluster)

	// Store the loaded cluster
	m.cluster = cluster

	return cluster, nil
}

// setClusterFromConfig applies all configuration values and defaults to the cluster.
func (m *Manager) setClusterFromConfig(cluster *v1alpha1.Cluster) {
	// Set metadata defaults
	m.setMetadataFromConfig(cluster)
	
	// Set spec defaults
	m.setSpecFromConfig(cluster)
	
	// Set connection defaults
	m.setConnectionFromConfig(cluster)
}

// setMetadataFromConfig sets metadata values from configuration with defaults.
func (m *Manager) setMetadataFromConfig(cluster *v1alpha1.Cluster) {
	// Set name - try hierarchical first, then apply default
	if name := m.viper.GetString("metadata.name"); name != "" {
		cluster.Metadata.Name = name
	} else {
		cluster.Metadata.Name = "ksail-default"
	}
}

// setSpecFromConfig sets spec values from configuration with defaults.
func (m *Manager) setSpecFromConfig(cluster *v1alpha1.Cluster) {
	// Distribution Config
	if distConfig := m.viper.GetString("distribution-config"); distConfig != "" {
		// CLI flag or env var is set
		cluster.Spec.DistributionConfig = distConfig
	} else if fileDistConfig := m.viper.GetString("spec.distributionConfig"); fileDistConfig != "" {
		// Config file is set
		cluster.Spec.DistributionConfig = fileDistConfig
	} else {
		cluster.Spec.DistributionConfig = "kind.yaml"
	}

	// Source Directory
	if sourceDir := m.viper.GetString("source-directory"); sourceDir != "" {
		// CLI flag or env var is set
		cluster.Spec.SourceDirectory = sourceDir
	} else if fileSourceDir := m.viper.GetString("spec.sourceDirectory"); fileSourceDir != "" {
		// Config file is set
		cluster.Spec.SourceDirectory = fileSourceDir
	} else {
		cluster.Spec.SourceDirectory = "k8s"
	}

	// Distribution - check CLI flag first, then config file, then default
	if distStr := m.viper.GetString("distribution"); distStr != "" {
		// CLI flag or env var is set
		var distribution v1alpha1.Distribution
		if err := distribution.Set(distStr); err == nil {
			cluster.Spec.Distribution = distribution
		} else {
			cluster.Spec.Distribution = v1alpha1.DistributionKind
		}
	} else if fileDistStr := m.viper.GetString("spec.distribution"); fileDistStr != "" {
		// Config file is set
		var distribution v1alpha1.Distribution
		if err := distribution.Set(fileDistStr); err == nil {
			cluster.Spec.Distribution = distribution
		} else {
			cluster.Spec.Distribution = v1alpha1.DistributionKind
		}
	} else {
		cluster.Spec.Distribution = v1alpha1.DistributionKind
	}

	// Reconciliation Tool
	if tool := m.viper.GetString("reconciliation-tool"); tool != "" {
		// CLI flag or env var is set
		var reconciliationTool v1alpha1.ReconciliationTool
		if err := reconciliationTool.Set(tool); err == nil {
			cluster.Spec.ReconciliationTool = reconciliationTool
		} else {
			cluster.Spec.ReconciliationTool = v1alpha1.ReconciliationToolKubectl
		}
	} else if fileTool := m.viper.GetString("spec.reconciliationTool"); fileTool != "" {
		// Config file is set
		var reconciliationTool v1alpha1.ReconciliationTool
		if err := reconciliationTool.Set(fileTool); err == nil {
			cluster.Spec.ReconciliationTool = reconciliationTool
		} else {
			cluster.Spec.ReconciliationTool = v1alpha1.ReconciliationToolKubectl
		}
	} else {
		cluster.Spec.ReconciliationTool = v1alpha1.ReconciliationToolKubectl
	}

	// CNI
	if cni := m.viper.GetString("c-n-i"); cni != "" {
		// CLI flag or env var is set
		cluster.Spec.CNI = v1alpha1.CNI(cni)
	} else if fileCni := m.viper.GetString("spec.cni"); fileCni != "" {
		// Config file is set
		cluster.Spec.CNI = v1alpha1.CNI(fileCni)
	} else {
		cluster.Spec.CNI = v1alpha1.CNIDefault
	}

	// CSI
	if csi := m.viper.GetString("c-s-i"); csi != "" {
		// CLI flag or env var is set
		cluster.Spec.CSI = v1alpha1.CSI(csi)
	} else if fileCSI := m.viper.GetString("spec.csi"); fileCSI != "" {
		// Config file is set
		cluster.Spec.CSI = v1alpha1.CSI(fileCSI)
	} else {
		cluster.Spec.CSI = v1alpha1.CSIDefault
	}

	// Ingress Controller
	if ingress := m.viper.GetString("ingress-controller"); ingress != "" {
		// CLI flag or env var is set
		cluster.Spec.IngressController = v1alpha1.IngressController(ingress)
	} else if fileIngress := m.viper.GetString("spec.ingressController"); fileIngress != "" {
		// Config file is set
		cluster.Spec.IngressController = v1alpha1.IngressController(fileIngress)
	} else {
		cluster.Spec.IngressController = v1alpha1.IngressControllerDefault
	}

	// Gateway Controller
	if gateway := m.viper.GetString("gateway-controller"); gateway != "" {
		// CLI flag or env var is set
		cluster.Spec.GatewayController = v1alpha1.GatewayController(gateway)
	} else if fileGateway := m.viper.GetString("spec.gatewayController"); fileGateway != "" {
		// Config file is set
		cluster.Spec.GatewayController = v1alpha1.GatewayController(fileGateway)
	} else {
		cluster.Spec.GatewayController = v1alpha1.GatewayControllerDefault
	}
}

const defaultConnectionTimeoutMinutes = 5

// setConnectionFromConfig sets connection values from configuration with defaults.
func (m *Manager) setConnectionFromConfig(cluster *v1alpha1.Cluster) {
	// Kubeconfig
	if kubeconfig := m.viper.GetString("connection-kubeconfig"); kubeconfig != "" {
		// CLI flag or env var is set
		cluster.Spec.Connection.Kubeconfig = kubeconfig
	} else if fileKubeconfig := m.viper.GetString("spec.connection.kubeconfig"); fileKubeconfig != "" {
		// Config file is set
		cluster.Spec.Connection.Kubeconfig = fileKubeconfig
	} else {
		cluster.Spec.Connection.Kubeconfig = "~/.kube/config"
	}

	// Context
	if context := m.viper.GetString("connection-context"); context != "" {
		// CLI flag or env var is set
		cluster.Spec.Connection.Context = context
	} else if fileContext := m.viper.GetString("spec.connection.context"); fileContext != "" {
		// Config file is set
		cluster.Spec.Connection.Context = fileContext
	} else {
		cluster.Spec.Connection.Context = "kind-ksail-default"
	}

	// Timeout
	if timeoutStr := m.viper.GetString("connection-timeout"); timeoutStr != "" {
		// CLI flag or env var is set
		if duration, err := time.ParseDuration(timeoutStr); err == nil {
			cluster.Spec.Connection.Timeout = metav1.Duration{Duration: duration}
		} else {
			cluster.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
		}
	} else if fileTimeoutStr := m.viper.GetString("spec.connection.timeout"); fileTimeoutStr != "" {
		// Config file is set
		if duration, err := time.ParseDuration(fileTimeoutStr); err == nil {
			cluster.Spec.Connection.Timeout = metav1.Duration{Duration: duration}
		} else {
			cluster.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
		}
	} else {
		cluster.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
	}
}

// GetCluster returns the currently loaded cluster configuration.
func (m *Manager) GetCluster() *v1alpha1.Cluster {
	if m.cluster == nil {
		// Load and return a default cluster using the config manager
		cluster, _ := m.LoadCluster()
		return cluster
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