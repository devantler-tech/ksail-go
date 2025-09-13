// Package config provides centralized configuration management using Viper.
// This file contains the main configuration Manager for handling cluster configuration.
package config

import (
	"reflect"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manager provides configuration management functionality using the v1alpha1.Cluster structure.
type Manager struct {
	viper          *viper.Viper
	cluster        *v1alpha1.Cluster
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	// testErrorHook is used for testing error scenarios - if set, LoadCluster will return this error
	testErrorHook error
}

// NewManager creates a new configuration manager.
// Field selectors are optional - if none provided, manager works for commands without configuration needs.
func NewManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *Manager {
	viperInstance := initializeViper()

	// Only set Viper defaults for the fields specified by field selectors
	// This achieves true zero-maintenance by reusing the same field selectors
	if len(fieldSelectors) > 0 {
		setViperDefaultsFromFieldSelectors(viperInstance, fieldSelectors)
	}

	return &Manager{
		viper:          viperInstance,
		cluster:        nil,
		fieldSelectors: fieldSelectors,
	}
}

// LoadCluster loads the cluster configuration from files and environment variables.
func (m *Manager) LoadCluster() (*v1alpha1.Cluster, error) {
	// Test hook for error scenarios
	if m.testErrorHook != nil {
		return nil, m.testErrorHook
	}

	// Start with a minimal cluster structure
	cluster := v1alpha1.NewCluster()

	// Apply configuration values from Viper (CLI flags, env vars, config files, field selector defaults)
	// Only fields that have been configured in Viper will be overridden
	m.setClusterFromViperConfig(cluster)

	// Store the loaded cluster
	m.cluster = cluster

	return cluster, nil
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

// SetTestErrorHook sets an error that LoadCluster will return - for testing only.
func (m *Manager) SetTestErrorHook(err error) {
	m.testErrorHook = err
}

// setViperDefaultsFromFieldSelectors sets Viper defaults for only the fields specified by field selectors.
// This reuses the same field selectors that commands provide for CLI flags to also set Viper defaults,
// eliminating the need for manual field mappings.
func setViperDefaultsFromFieldSelectors(
	viperInstance *viper.Viper,
	fieldSelectors []FieldSelector[v1alpha1.Cluster],
) {
	// Create a reference cluster for path discovery
	ref := v1alpha1.NewCluster()

	for _, fieldSelector := range fieldSelectors {
		processFieldSelector(viperInstance, ref, fieldSelector)
	}
}

// processFieldSelector processes a single field selector for Viper default setting.
func processFieldSelector(
	viperInstance *viper.Viper,
	ref *v1alpha1.Cluster,
	fieldSelector FieldSelector[v1alpha1.Cluster],
) {
	// Get the field reference from the selector
	fieldPtr := fieldSelector.selector(ref)
	if fieldPtr == nil {
		return
	}

	// Get the path dynamically from the field selector
	path := getFieldPathFromPointer(fieldPtr, ref)
	if path == "" {
		return
	}

	// Get the default value from the field selector
	defaultValue := fieldSelector.defaultValue
	if defaultValue == nil {
		return
	}

	// Convert typed default value to appropriate format for Viper
	viperValue := convertDefaultValueForViper(defaultValue)
	viperInstance.SetDefault(path, viperValue)
}

// convertDefaultValueForViper converts typed default values to appropriate format for Viper.
func convertDefaultValueForViper(defaultValue any) any {
	switch val := defaultValue.(type) {
	case v1alpha1.Distribution:
		return string(val)
	case v1alpha1.ReconciliationTool:
		return string(val)
	case v1alpha1.CNI:
		return string(val)
	case v1alpha1.CSI:
		return string(val)
	case v1alpha1.IngressController:
		return string(val)
	case v1alpha1.GatewayController:
		return string(val)
	case metav1.Duration:
		// Convert metav1.Duration to time.Duration for Viper
		return val.Duration
	default:
		return val
	}
}

// setClusterFromViperConfig applies configuration values to the cluster.
// This overlays configuration values from Viper onto the cluster that already has defaults.
func (m *Manager) setClusterFromViperConfig(cluster *v1alpha1.Cluster) {
	if len(m.fieldSelectors) > 0 {
		// Apply configuration for the specified field selectors
		// This achieves true zero-maintenance by reusing the same field selectors
		m.setClusterFromFieldSelectors(cluster)
	}
	// No fallback - commands must specify their field requirements
}

// setClusterFromFieldSelectors applies configuration values using the specified field selectors.
// This reuses the same field selectors that commands provide for CLI flags.
func (m *Manager) setClusterFromFieldSelectors(cluster *v1alpha1.Cluster) {
	// Create a reference cluster for path discovery
	ref := v1alpha1.NewCluster()

	for _, fieldSelector := range m.fieldSelectors {
		// Get the field reference from the selector using the reference cluster for path resolution
		refFieldPtr := fieldSelector.selector(ref)
		if refFieldPtr == nil {
			continue
		}

		// Get the path from the field pointer in the reference cluster
		path := getFieldPathFromPointer(refFieldPtr, ref)
		if path == "" {
			continue
		}

		// Get value from Viper
		value := m.getTypedValueFromViperByPath(path, cluster)

		// Get the corresponding field pointer in the target cluster
		targetFieldPtr := fieldSelector.selector(cluster)
		if targetFieldPtr != nil {
			m.setValueAtFieldPointer(cluster, targetFieldPtr, value)
		}
	}
}

// setValueAtFieldPointer sets a value at the field location specified by the field pointer.
func (m *Manager) setValueAtFieldPointer(_ *v1alpha1.Cluster, fieldPtr any, value any) {
	// Use reflection to set the value directly
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr || !fieldVal.Elem().CanSet() {
		return
	}

	// Convert and set the value
	targetField := fieldVal.Elem()

	convertedValue := convertValueToFieldType(value, targetField.Type())
	if convertedValue != nil {
		targetField.Set(reflect.ValueOf(convertedValue))
	}
}

// getTypedValueFromViperByPath retrieves a properly typed value from Viper based on the path.
func (m *Manager) getTypedValueFromViperByPath(path string, cluster *v1alpha1.Cluster) any {
	// Find the field type by using the path to navigate to the field
	targetFieldPtr := getFieldByPath(cluster, path)
	if targetFieldPtr == nil {
		// Fallback to string for unknown paths
		return m.viper.GetString(path)
	}

	fieldVal := reflect.ValueOf(targetFieldPtr)
	if fieldVal.Kind() != reflect.Ptr {
		return m.viper.GetString(path)
	}

	fieldType := fieldVal.Elem().Type()

	// Get the value from Viper based on the field type
	rawValue := m.getValueFromViperByType(path, fieldType)

	// Convert the value to the appropriate type
	return convertValueToFieldType(rawValue, fieldType)
}

// getValueFromViperByType retrieves a value from Viper using the appropriate method based on the field type.
func (m *Manager) getValueFromViperByType(path string, fieldType reflect.Type) any {
	// Handle boolean types
	if fieldType == reflect.TypeOf(true) {
		return m.viper.GetBool(path)
	}

	// Handle integer types
	if value := m.getIntegerValueFromViper(path, fieldType); value != nil {
		return value
	}

	// Handle float types
	if value := m.getFloatValueFromViper(path, fieldType); value != nil {
		return value
	}

	// Handle duration types
	if value := m.getDurationValueFromViper(path, fieldType); value != nil {
		return value
	}

	// Handle slice types
	if value := m.getSliceValueFromViper(path, fieldType); value != nil {
		return value
	}

	// For all other types (including custom types), get as string
	// The conversion function will handle the proper type conversion
	return m.viper.GetString(path)
}

// getIntegerValueFromViper handles integer type values from Viper.
func (m *Manager) getIntegerValueFromViper(path string, fieldType reflect.Type) any {
	switch fieldType {
	case reflect.TypeOf(int(0)):
		return m.viper.GetInt(path)
	case reflect.TypeOf(int32(0)):
		return m.viper.GetInt32(path)
	case reflect.TypeOf(int64(0)):
		return m.viper.GetInt64(path)
	case reflect.TypeOf(uint(0)):
		return m.viper.GetUint(path)
	case reflect.TypeOf(uint32(0)):
		return m.viper.GetUint32(path)
	case reflect.TypeOf(uint64(0)):
		return m.viper.GetUint64(path)
	}

	return nil
}

// getFloatValueFromViper handles float type values from Viper.
func (m *Manager) getFloatValueFromViper(path string, fieldType reflect.Type) any {
	switch fieldType {
	case reflect.TypeOf(float32(0)):
		return float32(m.viper.GetFloat64(path))
	case reflect.TypeOf(float64(0)):
		return m.viper.GetFloat64(path)
	}

	return nil
}

// getDurationValueFromViper handles duration type values from Viper.
func (m *Manager) getDurationValueFromViper(path string, fieldType reflect.Type) any {
	switch fieldType {
	case reflect.TypeOf(time.Duration(0)),
		reflect.TypeOf(metav1.Duration{Duration: time.Duration(0)}):
		return m.viper.GetDuration(path)
	}

	return nil
}

// getSliceValueFromViper handles slice type values from Viper.
func (m *Manager) getSliceValueFromViper(path string, fieldType reflect.Type) any {
	switch fieldType {
	case reflect.TypeOf([]string{}):
		return m.viper.GetStringSlice(path)
	case reflect.TypeOf([]int{}):
		return m.viper.GetIntSlice(path)
	}

	return nil
}
