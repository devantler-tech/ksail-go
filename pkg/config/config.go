// Package config provides centralized configuration management using Viper.
package config

import (
	"reflect"
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

// AutoBindFlags automatically binds CLI flags to Viper based on the v1alpha1.Cluster structure.
// This eliminates the need for manual BindPFlag calls for each flag.
func (m *Manager) AutoBindFlags(cmd *cobra.Command) {
	m.bindStructFlags(cmd, reflect.TypeOf(v1alpha1.Cluster{}), "")
	
	// Add common CLI-only flags that are not part of the cluster structure
	m.bindCLIOnlyFlags(cmd)
}

// BindSelectiveFlags binds only the specified flags.
func (m *Manager) BindSelectiveFlags(cmd *cobra.Command, flagsToInclude map[string]bool) {
	// Bind cluster structure flags selectively
	m.bindStructFlagsSelectively(cmd, reflect.TypeOf(v1alpha1.Cluster{}), "", flagsToInclude)
	
	// Bind CLI-only flags selectively
	m.bindCLIOnlyFlagsSelectively(cmd, flagsToInclude)
}

// bindCLIOnlyFlags adds CLI-specific flags that are not part of the cluster configuration.
func (m *Manager) bindCLIOnlyFlags(cmd *cobra.Command) {
	// Add 'all' flag for commands that need it
	cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
	_ = m.viper.BindPFlag("all", cmd.Flags().Lookup("all"))
}

// bindStructFlagsSelectively recursively binds only specified struct fields as CLI flags.
func (m *Manager) bindStructFlagsSelectively(cmd *cobra.Command, t reflect.Type, prefix string, flagsToInclude map[string]bool) {
	// Handle pointers and nested types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields and embedded metadata fields
		if !field.IsExported() || 
		   field.Name == "TypeMeta" || 
		   field.Name == "Metadata" ||
		   field.Type == reflect.TypeOf(metav1.ObjectMeta{}) {
			continue
		}

		// Get field name for the flag
		flagName := m.getFieldFlagName(field, prefix)
		if flagName == "" {
			continue
		}

		// Only bind if this flag is requested
		if !flagsToInclude[flagName] {
			// For nested structs, check if any nested flags are requested
			if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(metav1.Duration{}) {
				nestedPrefix := flagName
				if prefix != "" {
					nestedPrefix = prefix + "." + flagName
				}
				m.bindStructFlagsSelectively(cmd, field.Type, nestedPrefix, flagsToInclude)
			}
			continue
		}

		// Handle different field types (same logic as bindStructFlags)
		switch field.Type.Kind() {
		case reflect.String:
			cmd.Flags().String(flagName, "", m.getFieldDescription(field))
			_ = m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
			
		case reflect.Bool:
			cmd.Flags().Bool(flagName, false, m.getFieldDescription(field))
			_ = m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
			
		case reflect.Struct:
			if field.Type == reflect.TypeOf(metav1.Duration{}) {
				cmd.Flags().String(flagName, "", m.getFieldDescription(field))
				_ = m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
			} else {
				nestedPrefix := flagName
				if prefix != "" {
					nestedPrefix = prefix + "." + flagName
				}
				m.bindStructFlagsSelectively(cmd, field.Type, nestedPrefix, flagsToInclude)
			}
		}
	}
}

// bindCLIOnlyFlagsSelectively adds only requested CLI-specific flags.
func (m *Manager) bindCLIOnlyFlagsSelectively(cmd *cobra.Command, flagsToInclude map[string]bool) {
	if flagsToInclude["all"] {
		cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
		_ = m.viper.BindPFlag("all", cmd.Flags().Lookup("all"))
	}
}

// bindStructFlags recursively binds struct fields as CLI flags.
func (m *Manager) bindStructFlags(cmd *cobra.Command, t reflect.Type, prefix string) {
	// Handle pointers and nested types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields and embedded metadata fields
		if !field.IsExported() || 
		   field.Name == "TypeMeta" || 
		   field.Name == "Metadata" ||
		   field.Type == reflect.TypeOf(metav1.ObjectMeta{}) {
			continue
		}

		// Special handling for the Spec field - don't add it to the prefix, just recurse into it
		if field.Name == "Spec" && field.Type.Kind() == reflect.Struct {
			m.bindStructFlags(cmd, field.Type, prefix)
			continue
		}

		// Get field name for the flag
		flagName := m.getFieldFlagName(field, prefix)
		if flagName == "" {
			continue
		}

		// Handle different field types
		switch field.Type.Kind() {
		case reflect.String:
			// Handle string enums and regular strings
			if m.isEnumType(field.Type) {
				cmd.Flags().String(flagName, "", m.getFieldDescription(field))
			} else {
				cmd.Flags().String(flagName, "", m.getFieldDescription(field))
			}
			_ = m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
			
		case reflect.Bool:
			cmd.Flags().Bool(flagName, false, m.getFieldDescription(field))
			_ = m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
			
		case reflect.Struct:
			// Skip metav1.Duration and other special types, handle as strings
			if field.Type == reflect.TypeOf(metav1.Duration{}) {
				cmd.Flags().String(flagName, "", m.getFieldDescription(field))
				_ = m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
			} else {
				// Recursively handle nested structs
				nestedPrefix := flagName
				if prefix != "" {
					nestedPrefix = prefix + "-" + flagName
				}
				m.bindStructFlags(cmd, field.Type, nestedPrefix)
			}
		}
	}
}

// getFieldFlagName converts a struct field to a CLI flag name.
func (m *Manager) getFieldFlagName(field reflect.StructField, prefix string) string {
	// Convert camelCase to kebab-case
	flagName := m.camelToKebab(field.Name)
	
	// Add prefix for nested fields
	if prefix != "" {
		flagName = prefix + "-" + flagName
	}
	
	return flagName
}

// getFieldDescription generates a description for the CLI flag.
func (m *Manager) getFieldDescription(field reflect.StructField) string {
	// You could add struct tags for descriptions, for now use field name
	return "Configure " + strings.ToLower(field.Name)
}

// isEnumType checks if a type is a custom enum type (like Distribution).
func (m *Manager) isEnumType(t reflect.Type) bool {
	// Check if it's a custom string type (enum)
	return t.Kind() == reflect.String && t.PkgPath() != ""
}

// camelToKebab converts camelCase to kebab-case.
func (m *Manager) camelToKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
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