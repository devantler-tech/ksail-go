// Package config provides centralized configuration management using Viper.
package config

import (
	"reflect"
	"strings"

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