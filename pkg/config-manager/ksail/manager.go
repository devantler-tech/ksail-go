// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
package ksail

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FieldSelector defines a field and its metadata for configuration management.
type FieldSelector[T any] struct {
	Selector     func(*T) any // Function that returns a pointer to the field
	Description  string       // Human-readable description for CLI flags
	DefaultValue any          // Default value for the field
}

// Manager implements the ConfigManager interface for KSail v1alpha1.Cluster configurations.
type Manager struct {
	viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
}

// Verify that Manager implements the ConfigManager interface.
var _ configmanager.ConfigManager[v1alpha1.Cluster] = (*Manager)(nil)

// SuggestionsMinimumDistance represents the minimum levenshtein distance for command suggestions.
const SuggestionsMinimumDistance = 2

// NewManager creates a new configuration manager with the specified field selectors.
func NewManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *Manager {
	return &Manager{
		viper:          viper.New(),
		fieldSelectors: fieldSelectors,
		Config:         &v1alpha1.Cluster{},
	}
}

// AddFlagFromField returns a type-safe field selector for the given field path.
// This provides compile-time safety - if the struct changes, this will cause compilation errors.
// Requires a default value as the second parameter, optionally accepts a description as the third parameter.
//
// Usage:
//
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, v1alpha1.DistributionKind)
//	AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//		v1alpha1.DistributionKind, "Custom description")
func AddFlagFromField(
	selector func(*v1alpha1.Cluster) any,
	defaultValue any,
	description ...string,
) FieldSelector[v1alpha1.Cluster] {
	desc := ""
	if len(description) > 0 {
		desc = description[0]
	}

	return FieldSelector[v1alpha1.Cluster]{
		Selector:     selector,
		Description:  desc,
		DefaultValue: defaultValue,
	}
}

// NewCobraCommand creates a cobra.Command with automatic type-safe configuration binding.
// This is the only constructor provided for initializing CobraCommands.
// The binding automatically handles CLI flags (priority 1), environment variables (priority 2),
// configuration files (priority 3), and field selector defaults (priority 4).
//
// If fieldSelectors is provided, only those specific fields will be bound as CLI flags.
// Field selectors must include default values and optionally descriptions.
// If fieldSelectors is empty, no configuration flags will be added (no auto-discovery by default).
//
// Usage examples:
//
//	// No configuration flags (default behavior):
//	NewCobraCommand("status", "Show status", "...", handleStatusRunE)
//
//	// Type-safe selective binding with defaults and descriptions:
//	NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//	        v1alpha1.DistributionKind, "Kubernetes distribution to use"),
//	    AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
//	        "k8s", "Directory containing workloads to deploy"))
func NewCobraCommand(
	use, short, long string,
	runE func(*cobra.Command, *Manager, []string) error,
	fieldSelectors ...FieldSelector[v1alpha1.Cluster],
) *cobra.Command {
	manager := NewManager(fieldSelectors...)

	// Create the base command
	cmd := &cobra.Command{ //nolint:exhaustruct // Only setting needed fields
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
		// Bind only the specified field selectors for CLI flags
		manager.AddFlagsFromFields(cmd)
	}
	// No else clause - when no field selectors provided, no configuration flags are added

	return cmd
}

// LoadConfig loads the configuration from files and environment variables.
// Returns the previously loaded config if already loaded.
func (m *Manager) LoadConfig() (*v1alpha1.Cluster, error) {
	// If config is already loaded and populated, return it
	if m.Config != nil && !reflect.DeepEqual(m.Config, &v1alpha1.Cluster{}) {
		return m.Config, nil
	}

	// Initialize with defaults from field selectors
	m.applyDefaults()

	// Try to read from configuration files
	m.viper.SetConfigName("ksail")
	m.viper.SetConfigType("yaml")
	m.viper.AddConfigPath(".")
	m.viper.AddConfigPath("$HOME/.config/ksail")
	m.viper.AddConfigPath("/etc/ksail")

	// Read configuration file if it exists
	if err := m.viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and flags
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Set environment variable prefix
	m.viper.SetEnvPrefix("KSAIL")
	m.viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	m.viper.AutomaticEnv()

	// Bind environment variables to proper viper keys
	m.bindEnvironmentVariables()

	// Unmarshal into our cluster config
	if err := m.viper.Unmarshal(m.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return m.Config, nil
}

// GetViper returns the underlying Viper instance for flag binding.
func (m *Manager) GetViper() *viper.Viper {
	return m.viper
}

// AddFlagsFromFields adds CLI flags for all configured field selectors.
func (m *Manager) AddFlagsFromFields(cmd *cobra.Command) {
	for _, fieldSelector := range m.fieldSelectors {
		m.addFlagFromField(cmd, fieldSelector)
	}
}

// bindEnvironmentVariables binds environment variables to their corresponding viper keys.
func (m *Manager) bindEnvironmentVariables() {
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
		m.viper.BindEnv(viperKey, "KSAIL_"+envKey)
	}
}

// applyDefaults applies default values from field selectors to the config.
func (m *Manager) applyDefaults() {
	for _, fieldSelector := range m.fieldSelectors {
		fieldPtr := fieldSelector.Selector(m.Config)
		if fieldPtr != nil {
			m.setFieldValue(fieldPtr, fieldSelector.DefaultValue)
		}
	}
}

// addFlagFromField adds a CLI flag for a single field selector.
func (m *Manager) addFlagFromField(
	cmd *cobra.Command,
	fieldSelector FieldSelector[v1alpha1.Cluster],
) {
	fieldPtr := fieldSelector.Selector(m.Config)
	if fieldPtr == nil {
		return
	}

	flagName := m.generateFlagName(fieldPtr)
	shorthand := m.generateShorthand(flagName)

	switch ptr := fieldPtr.(type) {
	case *v1alpha1.Distribution:
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, string(fieldSelector.DefaultValue.(v1alpha1.Distribution)), fieldSelector.Description)
	case *v1alpha1.ReconciliationTool:
		defaultTool := ""
		if fieldSelector.DefaultValue != nil {
			defaultTool = string(fieldSelector.DefaultValue.(v1alpha1.ReconciliationTool))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultTool, fieldSelector.Description)
	case *v1alpha1.CNI:
		defaultCNI := ""
		if fieldSelector.DefaultValue != nil {
			defaultCNI = string(fieldSelector.DefaultValue.(v1alpha1.CNI))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultCNI, fieldSelector.Description)
	case *v1alpha1.CSI:
		defaultCSI := ""
		if fieldSelector.DefaultValue != nil {
			defaultCSI = string(fieldSelector.DefaultValue.(v1alpha1.CSI))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultCSI, fieldSelector.Description)
	case *v1alpha1.IngressController:
		defaultIngress := ""
		if fieldSelector.DefaultValue != nil {
			defaultIngress = string(fieldSelector.DefaultValue.(v1alpha1.IngressController))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultIngress, fieldSelector.Description)
	case *v1alpha1.GatewayController:
		defaultGateway := ""
		if fieldSelector.DefaultValue != nil {
			defaultGateway = string(fieldSelector.DefaultValue.(v1alpha1.GatewayController))
		}
		cmd.Flags().StringVarP((*string)(ptr), flagName, shorthand, defaultGateway, fieldSelector.Description)
	case *string:
		defaultStr := ""
		if fieldSelector.DefaultValue != nil {
			defaultStr = fieldSelector.DefaultValue.(string)
		}
		cmd.Flags().StringVarP(ptr, flagName, shorthand, defaultStr, fieldSelector.Description)
	case *bool:
		defaultBool := false
		if fieldSelector.DefaultValue != nil {
			defaultBool = fieldSelector.DefaultValue.(bool)
		}
		cmd.Flags().BoolVarP(ptr, flagName, shorthand, defaultBool, fieldSelector.Description)
	case *int:
		defaultInt := 0
		if fieldSelector.DefaultValue != nil {
			defaultInt = fieldSelector.DefaultValue.(int)
		}
		cmd.Flags().IntVarP(ptr, flagName, shorthand, defaultInt, fieldSelector.Description)
	case *metav1.Duration:
		defaultDuration := time.Duration(0)
		if fieldSelector.DefaultValue != nil {
			if dur, ok := fieldSelector.DefaultValue.(metav1.Duration); ok {
				defaultDuration = dur.Duration
			}
		}
		cmd.Flags().DurationVarP(&ptr.Duration, flagName, shorthand, defaultDuration, fieldSelector.Description)
	case *time.Duration:
		defaultDuration := time.Duration(0)
		if fieldSelector.DefaultValue != nil {
			defaultDuration = fieldSelector.DefaultValue.(time.Duration)
		}
		cmd.Flags().DurationVarP(ptr, flagName, shorthand, defaultDuration, fieldSelector.Description)
	}

	// Bind the flag to viper
	if err := m.viper.BindPFlag(flagName, cmd.Flags().Lookup(flagName)); err != nil {
		// Log error but don't fail - this is not critical
		fmt.Printf("Warning: failed to bind flag %s to viper: %v\n", flagName, err)
	}
}

// generateFlagName generates a user-friendly flag name from a field pointer.
func (m *Manager) generateFlagName(fieldPtr any) string {
	// Check which field this pointer references by comparing addresses
	if fieldPtr == &m.Config.Spec.Distribution {
		return "distribution"
	}
	if fieldPtr == &m.Config.Spec.DistributionConfig {
		return "distribution-config"
	}
	if fieldPtr == &m.Config.Spec.SourceDirectory {
		return "source-directory"
	}
	if fieldPtr == &m.Config.Spec.Connection.Context {
		return "context"
	}
	if fieldPtr == &m.Config.Spec.Connection.Kubeconfig {
		return "kubeconfig"
	}
	if fieldPtr == &m.Config.Spec.Connection.Timeout {
		return "timeout"
	}
	if fieldPtr == &m.Config.Spec.ReconciliationTool {
		return "reconciliation-tool"
	}
	if fieldPtr == &m.Config.Spec.CNI {
		return "cni"
	}
	if fieldPtr == &m.Config.Spec.CSI {
		return "csi"
	}
	if fieldPtr == &m.Config.Spec.IngressController {
		return "ingress-controller"
	}
	if fieldPtr == &m.Config.Spec.GatewayController {
		return "gateway-controller"
	}

	// Fallback to field name detection for other fields
	fieldName := m.getFieldNameFromPointer(fieldPtr, m.Config)
	return strings.ToLower(fieldName)
}

// getFieldNameFromPointer finds the field name for a given pointer.
func (m *Manager) getFieldNameFromPointer(fieldPtr any, rootStruct any) string {
	return m.findFieldName(reflect.ValueOf(fieldPtr), reflect.ValueOf(rootStruct))
}

// findFieldName recursively finds the field name for a pointer.
func (m *Manager) findFieldName(targetPtr, current reflect.Value) string {
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return ""
		}
		current = current.Elem()
	}

	if current.Kind() != reflect.Struct {
		return ""
	}

	typ := current.Type()
	for i := 0; i < current.NumField(); i++ {
		field := current.Field(i)
		fieldType := typ.Field(i)

		if !field.CanAddr() {
			continue
		}

		fieldAddr := field.Addr()
		if fieldAddr.Pointer() == targetPtr.Pointer() {
			return fieldType.Name
		}

		// Recurse into nested structs
		if field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			if result := m.findFieldName(targetPtr, field); result != "" {
				return result
			}
		}
	}

	return ""
}

// generateShorthand generates a shorthand flag from the flag name.
func (m *Manager) generateShorthand(flagName string) string {
	switch flagName {
	case "distribution":
		return "d"
	case "context":
		return "c"
	case "kubeconfig":
		return "k"
	case "timeout":
		return "t"
	case "source-directory":
		return "s"
	case "reconciliation-tool":
		return "r"
	case "distribution-config":
		return ""
	default:
		// For other flags, don't provide shorthand to avoid conflicts
		return ""
	}
}

// getFieldPath returns the JSON path of a field pointer within a struct.
func (m *Manager) getFieldPath(fieldPtr any, rootStruct any) string {
	return m.findFieldPath(reflect.ValueOf(fieldPtr), reflect.ValueOf(rootStruct), "")
}

// findFieldPath recursively finds the path to a field pointer.
func (m *Manager) findFieldPath(targetPtr, current reflect.Value, currentPath string) string {
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return ""
		}
		current = current.Elem()
	}

	if current.Kind() != reflect.Struct {
		return ""
	}

	typ := current.Type()
	for i := 0; i < current.NumField(); i++ {
		field := current.Field(i)
		fieldType := typ.Field(i)

		if !field.CanAddr() {
			continue
		}

		fieldAddr := field.Addr()
		if fieldAddr.Pointer() == targetPtr.Pointer() {
			fieldName := fieldType.Name
			if currentPath == "" {
				return fieldName
			}
			return currentPath + "." + fieldName
		}

		// Recurse into nested structs
		if field.Kind() == reflect.Struct ||
			(field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			newPath := fieldType.Name
			if currentPath != "" {
				newPath = currentPath + "." + newPath
			}
			if result := m.findFieldPath(targetPtr, field, newPath); result != "" {
				return result
			}
		}
	}

	return ""
}

// setFieldValue sets a field value using reflection.
func (m *Manager) setFieldValue(fieldPtr any, value any) {
	if fieldPtr == nil || value == nil {
		return
	}

	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr || fieldVal.IsNil() {
		return
	}

	fieldVal = fieldVal.Elem()
	valueVal := reflect.ValueOf(value)

	if fieldVal.Type().AssignableTo(valueVal.Type()) {
		fieldVal.Set(valueVal)
	} else if fieldVal.Type().ConvertibleTo(valueVal.Type()) {
		fieldVal.Set(valueVal.Convert(fieldVal.Type()))
	}
}
