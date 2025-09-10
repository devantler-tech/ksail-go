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
	defaultCluster *v1alpha1.Cluster
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
}

// NewManager creates a new configuration manager.
func NewManager() *Manager {
	v := initializeViper()
	defaultCluster := v1alpha1.NewDefaultCluster()

	// For backward compatibility, set all defaults when no field selectors are provided
	setViperDefaultsFromAllFields(v, defaultCluster)

	return &Manager{
		viper:          v,
		cluster:        nil,
		defaultCluster: defaultCluster,
		fieldSelectors: nil,
	}
}

// NewManagerWithFieldSelectors creates a new configuration manager with specific field selectors.
// This allows the manager to only handle configuration for fields that commands actually need.
func NewManagerWithFieldSelectors(fieldSelectors []FieldSelector[v1alpha1.Cluster]) *Manager {
	v := initializeViper()
	defaultCluster := v1alpha1.NewDefaultCluster()

	if len(fieldSelectors) > 0 {
		// Only set Viper defaults for the fields specified by field selectors
		// This achieves true zero-maintenance by reusing the same field selectors
		setViperDefaultsFromFieldSelectors(v, defaultCluster, fieldSelectors)
	} else {
		// When no field selectors are provided, fall back to all fields for backward compatibility
		setViperDefaultsFromAllFields(v, defaultCluster)
	}

	return &Manager{
		viper:          v,
		cluster:        nil,
		defaultCluster: defaultCluster,
		fieldSelectors: fieldSelectors,
	}
}

// LoadCluster loads the cluster configuration from files and environment variables.
func (m *Manager) LoadCluster() (*v1alpha1.Cluster, error) {
	// Start with a cluster that has all meaningful defaults from the constructor
	cluster := v1alpha1.NewDefaultCluster()

	// Apply configuration values from Viper (CLI flags, env vars, config files)
	// Only fields that have been configured in Viper will be overridden
	m.setClusterFromConfig(cluster)

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
	defaultCluster *v1alpha1.Cluster,
	fieldSelectors []FieldSelector[v1alpha1.Cluster],
) {
	for _, fieldSelector := range fieldSelectors {
		// Get the field reference from the selector
		fieldPtr := fieldSelector.selector(defaultCluster)
		if fieldPtr == nil {
			continue
		}

		// Get the path dynamically from the field selector
		path := getFieldPathFromPointer(fieldPtr, defaultCluster)
		if path == "" {
			continue
		}

		// Get the default value from the default cluster instance
		defaultValue := getValueFromFieldPointer(fieldPtr)
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
		default:
			viperValue = val
		}

		v.SetDefault(path, viperValue)
	}
}

// getValueFromFieldPointer extracts the value from a field pointer.
func getValueFromFieldPointer(fieldPtr any) any {
	fieldVal := reflect.ValueOf(fieldPtr)
	if fieldVal.Kind() != reflect.Ptr || !fieldVal.IsValid() {
		return nil
	}

	elem := fieldVal.Elem()
	if !elem.IsValid() {
		return nil
	}

	return elem.Interface()
}

// setViperDefaultsFromAllFields sets all configuration defaults in Viper for backward compatibility.
// This is used when no field selectors are provided.
func setViperDefaultsFromAllFields(v *viper.Viper, defaultCluster *v1alpha1.Cluster) {
	// Get all field selectors from the default cluster
	configDefaults := getConfigDefaultsFromCluster(defaultCluster)

	for _, configDefault := range configDefaults {
		// Get the path dynamically from the field selector
		path := getFieldPathFromPointer(configDefault.FieldPtr, defaultCluster)

		// Convert typed default value to appropriate format for Viper
		var viperValue any
		switch val := configDefault.DefaultValue.(type) {
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
		default:
			viperValue = val
		}

		v.SetDefault(path, viperValue)
	}
}

// configDefaultEntry represents a configuration field with its default value.
type configDefaultEntry struct {
	FieldPtr     any
	DefaultValue any
}

// getConfigDefaultsFromCluster extracts configuration defaults from a default cluster instance.
func getConfigDefaultsFromCluster(defaultCluster *v1alpha1.Cluster) []configDefaultEntry {
	return []configDefaultEntry{
		// Metadata defaults
		{
			FieldPtr:     &defaultCluster.Metadata.Name,
			DefaultValue: defaultCluster.Metadata.Name,
		},

		// Spec defaults
		{
			FieldPtr:     &defaultCluster.Spec.DistributionConfig,
			DefaultValue: defaultCluster.Spec.DistributionConfig,
		},
		{
			FieldPtr:     &defaultCluster.Spec.SourceDirectory,
			DefaultValue: defaultCluster.Spec.SourceDirectory,
		},
		{
			FieldPtr:     &defaultCluster.Spec.Distribution,
			DefaultValue: defaultCluster.Spec.Distribution,
		},
		{
			FieldPtr:     &defaultCluster.Spec.ReconciliationTool,
			DefaultValue: defaultCluster.Spec.ReconciliationTool,
		},
		{
			FieldPtr:     &defaultCluster.Spec.CNI,
			DefaultValue: defaultCluster.Spec.CNI,
		},
		{
			FieldPtr:     &defaultCluster.Spec.CSI,
			DefaultValue: defaultCluster.Spec.CSI,
		},
		{
			FieldPtr:     &defaultCluster.Spec.IngressController,
			DefaultValue: defaultCluster.Spec.IngressController,
		},
		{
			FieldPtr:     &defaultCluster.Spec.GatewayController,
			DefaultValue: defaultCluster.Spec.GatewayController,
		},

		// Connection defaults
		{
			FieldPtr:     &defaultCluster.Spec.Connection.Kubeconfig,
			DefaultValue: defaultCluster.Spec.Connection.Kubeconfig,
		},
		{
			FieldPtr:     &defaultCluster.Spec.Connection.Context,
			DefaultValue: defaultCluster.Spec.Connection.Context,
		},
		{
			FieldPtr:     &defaultCluster.Spec.Connection.Timeout,
			DefaultValue: defaultCluster.Spec.Connection.Timeout,
		},
	}
}

// setClusterFromConfig applies configuration values to the cluster.
// This overlays configuration values from Viper onto the cluster that already has defaults.
func (m *Manager) setClusterFromConfig(cluster *v1alpha1.Cluster) {
	if len(m.fieldSelectors) > 0 {
		// When field selectors are provided, only apply configuration for those specific fields
		// This achieves true zero-maintenance by reusing the same field selectors
		m.setClusterFromFieldSelectors(cluster)
	} else {
		// Fallback to applying all configuration for backward compatibility
		m.setClusterFromAllDefaults(cluster)
	}
}

// setClusterFromFieldSelectors applies configuration values using the specified field selectors.
// This reuses the same field selectors that commands provide for CLI flags.
func (m *Manager) setClusterFromFieldSelectors(cluster *v1alpha1.Cluster) {
	for _, fieldSelector := range m.fieldSelectors {
		// Get the field reference from the selector using the default cluster for path resolution
		defaultFieldPtr := fieldSelector.selector(m.defaultCluster)
		if defaultFieldPtr == nil {
			continue
		}

		// Get the path from the field pointer in the default cluster
		path := getFieldPathFromPointer(defaultFieldPtr, m.defaultCluster)
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

// setClusterFromAllDefaults applies all default configuration values.
func (m *Manager) setClusterFromAllDefaults(cluster *v1alpha1.Cluster) {
	// Get all field mappings from the default cluster
	configDefaults := getConfigDefaultsFromCluster(m.defaultCluster)

	for _, configDefault := range configDefaults {
		// Get the path from the field pointer
		path := getFieldPathFromPointer(configDefault.FieldPtr, m.defaultCluster)

		// Get value from Viper and set it in the cluster
		value := m.getTypedValueFromViperByPath(path, cluster)
		
		// Get the corresponding field pointer in the target cluster instance
		targetFieldPtr := getFieldByPath(cluster, path)
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
