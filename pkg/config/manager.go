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
}

// NewManager creates a new configuration manager.
// Field selectors are optional - if none provided, manager works for commands without configuration needs.
func NewManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *Manager {
	v := initializeViper()

	// Only set Viper defaults for the fields specified by field selectors
	// This achieves true zero-maintenance by reusing the same field selectors
	if len(fieldSelectors) > 0 {
		setViperDefaultsFromFieldSelectors(v, fieldSelectors)
	}

	return &Manager{
		viper:          v,
		cluster:        nil,
		fieldSelectors: fieldSelectors,
	}
}

// LoadCluster loads the cluster configuration from files and environment variables.
func (m *Manager) LoadCluster() (*v1alpha1.Cluster, error) {
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

// setViperDefaultsFromFieldSelectors sets Viper defaults for only the fields specified by field selectors.
// This reuses the same field selectors that commands provide for CLI flags to also set Viper defaults,
// eliminating the need for manual field mappings.
func setViperDefaultsFromFieldSelectors(
	v *viper.Viper,
	fieldSelectors []FieldSelector[v1alpha1.Cluster],
) {
	// Create a reference cluster for path discovery
	ref := v1alpha1.NewCluster()

	for _, fieldSelector := range fieldSelectors {
		// Get the field reference from the selector
		fieldPtr := fieldSelector.selector(ref)
		if fieldPtr == nil {
			continue
		}

		// Get the path dynamically from the field selector
		path := getFieldPathFromPointer(fieldPtr, ref)
		if path == "" {
			continue
		}

		// Get the default value from the field selector
		defaultValue := fieldSelector.defaultValue
		if defaultValue == nil {
			continue
		}

		// Convert typed default value to appropriate format for Viper
		var viperValue any
		switch val := defaultValue.(type) {
		case v1alpha1.Distribution:
			viperValue = string(val)
		case v1alpha1.ReconciliationTool:
			viperValue = string(val)
		case v1alpha1.CNI:
			viperValue = string(val)
		case v1alpha1.CSI:
			viperValue = string(val)
		case v1alpha1.IngressController:
			viperValue = string(val)
		case v1alpha1.GatewayController:
			viperValue = string(val)
		case metav1.Duration:
			// Convert metav1.Duration to time.Duration for Viper
			viperValue = val.Duration
		default:
			viperValue = val
		}

		v.SetDefault(path, viperValue)
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
func (m *Manager) setValueAtFieldPointer(cluster *v1alpha1.Cluster, fieldPtr any, value any) {
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
	switch fieldType {
	case reflect.TypeOf(true):
		return m.viper.GetBool(path)
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
	case reflect.TypeOf(float32(0)):
		return float32(m.viper.GetFloat64(path))
	case reflect.TypeOf(float64(0)):
		return m.viper.GetFloat64(path)
	case reflect.TypeOf(time.Duration(0)):
		return m.viper.GetDuration(path)
	case reflect.TypeOf([]string{}):
		return m.viper.GetStringSlice(path)
	case reflect.TypeOf([]int{}):
		return m.viper.GetIntSlice(path)
	case reflect.TypeOf(metav1.Duration{}):
		return m.viper.GetDuration(path)
	default:
		// For all other types (including custom types), get as string
		// The conversion function will handle the proper type conversion
		return m.viper.GetString(path)
	}
}
