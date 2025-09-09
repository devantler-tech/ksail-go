// Package config provides centralized configuration management using Viper.
package config

import (
	"fmt"
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
	// SuggestionsMinimumDistance is the minimum edit distance to suggest a command.
	SuggestionsMinimumDistance = 2
)

// FieldSelector represents a type-safe field selector for auto-binding.
// It provides compile-time safety by referencing actual struct fields.
type FieldSelector[T any] func(*T) any

// Predefined field selectors for common configuration fields.
// These provide compile-time safety - if the struct changes, these will cause compilation errors.
var (
	// Metadata fields
	MetadataNameField FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Metadata.Name }
	
	// Spec fields
	SpecDistributionField       FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }
	SpecDistributionConfigField FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig }
	SpecSourceDirectoryField    FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory }
	SpecCNIField                FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.CNI }
	SpecCSIField                FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.CSI }
	SpecIngressControllerField  FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.IngressController }
	SpecGatewayControllerField  FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.GatewayController }
	SpecReconciliationToolField FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool }
	
	// Connection fields
	SpecConnectionKubeconfigField FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig }
	SpecConnectionContextField    FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context }
	SpecConnectionTimeoutField    FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout }
	
	// Special CLI-only fields (not part of the cluster structure)
	AllField FieldSelector[v1alpha1.Cluster] = func(c *v1alpha1.Cluster) any { return nil } // Special marker for CLI-only flags
)


// NewCobraCommand creates a cobra.Command with automatic type-safe configuration binding.
// This is the only constructor provided for initializing CobraCommands with field selectors.
// The binding automatically handles CLI flags (priority 1), environment variables (priority 2), 
// and configuration defaults (priority 3).
func NewCobraCommand(
	use, short, long string,
	runE func(*cobra.Command, *Manager, []string) error,
	fieldSelectors []FieldSelector[v1alpha1.Cluster],
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
	
	// Auto-bind flags based on field selectors
	if len(fieldSelectors) > 0 {
		bindFieldSelectors(cmd, manager, fieldSelectors)
	}
	
	return cmd
}

// bindFieldSelectors automatically discovers and binds CLI flags for the specified field selectors.
func bindFieldSelectors(cmd *cobra.Command, manager *Manager, fieldSelectors []FieldSelector[v1alpha1.Cluster]) {
	// Create a dummy cluster to introspect field paths
	dummy := &v1alpha1.Cluster{}
	
	for _, selector := range fieldSelectors {
		// Get the field reference from the selector
		fieldPtr := selector(dummy)
		
		// Handle special CLI-only fields
		if fieldPtr == nil {
			// This is a special CLI-only field (like AllField)
			if isAllField(selector) {
				cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
				_ = manager.viper.BindPFlag("all", cmd.Flags().Lookup("all"))
			}
			continue
		}
		
		// Use reflection to discover the field path
		fieldPath := getFieldPath(dummy, fieldPtr)
		if fieldPath == "" {
			continue
		}
		
		// Convert hierarchical path to kebab-case CLI flag
		flagName := pathToFlagName(fieldPath)
		
		// Generate description
		description := generateFieldDescription(fieldPath)
		
		// Add string flag (all config values are treated as strings initially)
		cmd.Flags().String(flagName, "", description)
		
		// Bind to both the hierarchical path (for config files) and the flat flag name (for CLI/env)
		_ = manager.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName))
		_ = manager.viper.BindPFlag(fieldPath, cmd.Flags().Lookup(flagName))
	}
}

// isAllField checks if a field selector is the special AllField.
func isAllField(selector FieldSelector[v1alpha1.Cluster]) bool {
	// Compare function pointers - this is a Go-specific way to identify the specific function
	dummy := &v1alpha1.Cluster{}
	return selector(dummy) == nil && reflect.ValueOf(selector).Pointer() == reflect.ValueOf(AllField).Pointer()
}

// getFieldPath uses reflection to determine the path of a field within the cluster structure.
func getFieldPath(cluster *v1alpha1.Cluster, fieldPtr any) string {
	// Get the value and type of the cluster
	clusterVal := reflect.ValueOf(cluster).Elem()
	clusterType := clusterVal.Type()
	
	// Convert the field pointer to a reflect.Value
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr {
		return ""
	}
	fieldAddr := fieldVal.Pointer()
	
	// Recursively find the field path
	return findFieldPath(clusterVal, clusterType, fieldAddr, "")
}

// findFieldPath recursively searches for a field's path in a struct.
func findFieldPath(structVal reflect.Value, structType reflect.Type, targetAddr uintptr, prefix string) string {
	for i := 0; i < structVal.NumField(); i++ {
		field := structVal.Field(i)
		fieldType := structType.Field(i)
		
		// Skip unexported fields
		if !field.CanAddr() {
			continue
		}
		
		// Build the current field path
		var currentPath string
		if prefix == "" {
			currentPath = strings.ToLower(fieldType.Name)
		} else {
			currentPath = prefix + "." + strings.ToLower(fieldType.Name)
		}
		
		// Check if this field's address matches our target
		if field.CanAddr() && field.Addr().Pointer() == targetAddr {
			return currentPath
		}
		
		// If this is a struct, recurse into it
		if field.Kind() == reflect.Struct && !isTimeType(field.Type()) {
			if result := findFieldPath(field, field.Type(), targetAddr, currentPath); result != "" {
				return result
			}
		}
	}
	
	return ""
}

// isTimeType checks if a type is a time-related type that shouldn't be recursed into.
func isTimeType(t reflect.Type) bool {
	return t == reflect.TypeOf(time.Time{}) || t == reflect.TypeOf(metav1.Duration{})
}

// pathToFlagName converts a hierarchical field path to a kebab-case CLI flag name.
// E.g., "metadata.name" -> "metadata-name", "spec.connection.kubeconfig" -> "spec-connection-kubeconfig"
func pathToFlagName(path string) string {
	return strings.ReplaceAll(path, ".", "-")
}

// generateFieldDescription generates a human-readable description for a configuration field.
func generateFieldDescription(fieldPath string) string {
	switch fieldPath {
	case "metadata.name":
		return "Configure cluster name"
	case "spec.distribution":
		return "Configure cluster distribution (Kind, K3d, EKS, Tind)"
	case "spec.distributionconfig":
		return "Configure distribution config file path"
	case "spec.sourcedirectory":
		return "Configure source directory for Kubernetes manifests"
	case "spec.connection.kubeconfig":
		return "Configure kubeconfig file path"
	case "spec.connection.context":
		return "Configure kubectl context"
	case "spec.connection.timeout":
		return "Configure connection timeout duration"
	case "spec.cni":
		return "Configure CNI (Container Network Interface)"
	case "spec.csi":
		return "Configure CSI (Container Storage Interface)"
	case "spec.ingresscontroller":
		return "Configure ingress controller"
	case "spec.gatewaycontroller":
		return "Configure gateway controller"
	case "spec.reconciliationtool":
		return "Configure reconciliation tool (Kubectl, Flux, ArgoCD)"
	default:
		// Generate a default description based on the field path
		parts := strings.Split(fieldPath, ".")
		lastPart := parts[len(parts)-1]
		return fmt.Sprintf("Configure %s", strings.ReplaceAll(lastPart, "_", " "))
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
	if name := m.viper.GetString("metadata-name"); name != "" {
		cluster.Metadata.Name = name
	} else if name := m.viper.GetString("metadata.name"); name != "" {
		cluster.Metadata.Name = name
	} else {
		cluster.Metadata.Name = "ksail-default"
	}
}

// setSpecFromConfig sets spec values from configuration with defaults.
func (m *Manager) setSpecFromConfig(cluster *v1alpha1.Cluster) {
	// Distribution Config
	if distConfig := m.viper.GetString("spec-distributionconfig"); distConfig != "" {
		// CLI flag or env var is set
		cluster.Spec.DistributionConfig = distConfig
	} else if fileDistConfig := m.viper.GetString("spec.distributionconfig"); fileDistConfig != "" {
		// Config file is set
		cluster.Spec.DistributionConfig = fileDistConfig
	} else {
		cluster.Spec.DistributionConfig = "kind.yaml"
	}

	// Source Directory
	if sourceDir := m.viper.GetString("spec-sourcedirectory"); sourceDir != "" {
		// CLI flag or env var is set
		cluster.Spec.SourceDirectory = sourceDir
	} else if fileSourceDir := m.viper.GetString("spec.sourcedirectory"); fileSourceDir != "" {
		// Config file is set
		cluster.Spec.SourceDirectory = fileSourceDir
	} else {
		cluster.Spec.SourceDirectory = "k8s"
	}

	// Distribution - check CLI flag first, then config file, then default
	if distStr := m.viper.GetString("spec-distribution"); distStr != "" {
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
	if tool := m.viper.GetString("spec-reconciliationtool"); tool != "" {
		// CLI flag or env var is set
		var reconciliationTool v1alpha1.ReconciliationTool
		if err := reconciliationTool.Set(tool); err == nil {
			cluster.Spec.ReconciliationTool = reconciliationTool
		} else {
			cluster.Spec.ReconciliationTool = v1alpha1.ReconciliationToolKubectl
		}
	} else if fileTool := m.viper.GetString("spec.reconciliationtool"); fileTool != "" {
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
	if cni := m.viper.GetString("spec-cni"); cni != "" {
		// CLI flag or env var is set
		cluster.Spec.CNI = v1alpha1.CNI(cni)
	} else if fileCni := m.viper.GetString("spec.cni"); fileCni != "" {
		// Config file is set
		cluster.Spec.CNI = v1alpha1.CNI(fileCni)
	} else {
		cluster.Spec.CNI = v1alpha1.CNIDefault
	}

	// CSI
	if csi := m.viper.GetString("spec-csi"); csi != "" {
		// CLI flag or env var is set
		cluster.Spec.CSI = v1alpha1.CSI(csi)
	} else if fileCSI := m.viper.GetString("spec.csi"); fileCSI != "" {
		// Config file is set
		cluster.Spec.CSI = v1alpha1.CSI(fileCSI)
	} else {
		cluster.Spec.CSI = v1alpha1.CSIDefault
	}

	// Ingress Controller
	if ingress := m.viper.GetString("spec-ingresscontroller"); ingress != "" {
		// CLI flag or env var is set
		cluster.Spec.IngressController = v1alpha1.IngressController(ingress)
	} else if fileIngress := m.viper.GetString("spec.ingresscontroller"); fileIngress != "" {
		// Config file is set
		cluster.Spec.IngressController = v1alpha1.IngressController(fileIngress)
	} else {
		cluster.Spec.IngressController = v1alpha1.IngressControllerDefault
	}

	// Gateway Controller
	if gateway := m.viper.GetString("spec-gatewaycontroller"); gateway != "" {
		// CLI flag or env var is set
		cluster.Spec.GatewayController = v1alpha1.GatewayController(gateway)
	} else if fileGateway := m.viper.GetString("spec.gatewaycontroller"); fileGateway != "" {
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
	if kubeconfig := m.viper.GetString("spec-connection-kubeconfig"); kubeconfig != "" {
		// CLI flag or env var is set
		cluster.Spec.Connection.Kubeconfig = kubeconfig
	} else if fileKubeconfig := m.viper.GetString("spec.connection.kubeconfig"); fileKubeconfig != "" {
		// Config file is set
		cluster.Spec.Connection.Kubeconfig = fileKubeconfig
	} else {
		cluster.Spec.Connection.Kubeconfig = "~/.kube/config"
	}

	// Context
	if context := m.viper.GetString("spec-connection-context"); context != "" {
		// CLI flag or env var is set
		cluster.Spec.Connection.Context = context
	} else if fileContext := m.viper.GetString("spec.connection.context"); fileContext != "" {
		// Config file is set
		cluster.Spec.Connection.Context = fileContext
	} else {
		cluster.Spec.Connection.Context = "kind-ksail-default"
	}

	// Timeout
	if timeoutStr := m.viper.GetString("spec-connection-timeout"); timeoutStr != "" {
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