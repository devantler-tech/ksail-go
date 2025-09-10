package config

import (
	"reflect"
	"time"

	"github.com/devantler-tech/ksail-go/internal/utils/k8s"
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manager provides configuration management functionality using the v1alpha1.Cluster structure.
type Manager struct {
	viper          *viper.Viper
	cluster        *v1alpha1.Cluster
	defaultCluster *v1alpha1.Cluster
}

// NewManager creates a new configuration manager.
func NewManager() *Manager {
	v := initializeViper()
	defaultCluster := v1alpha1.NewDefaultCluster()
	
	// Set all defaults using the default cluster instance
	setViperDefaultsFromCluster(v, defaultCluster)
	
	return &Manager{
		viper:          v,
		cluster:        nil,
		defaultCluster: defaultCluster,
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

	// Apply all configuration and defaults through the default cluster
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

// setViperDefaultsFromCluster sets all configuration defaults in Viper using the default cluster instance.
func setViperDefaultsFromCluster(v *viper.Viper, defaultCluster *v1alpha1.Cluster) {
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

// setClusterFromConfig applies all configuration values to the cluster using the default cluster as reference.
func (m *Manager) setClusterFromConfig(cluster *v1alpha1.Cluster) {
	// Get all field mappings from the default cluster
	configDefaults := getConfigDefaultsFromCluster(m.defaultCluster)

	for _, configDefault := range configDefaults {
		// Get the path from the field pointer
		path := getFieldPathFromPointer(configDefault.FieldPtr, m.defaultCluster)
		
		// Get value from Viper and set it in the cluster
		value := m.getTypedValueFromViperByPath(path, cluster)
		setValueAtFieldPointer(cluster, configDefault.FieldPtr, m.defaultCluster, value)
	}
}

// setValueAtFieldPointer sets a value at the field location specified by the field pointer.
func setValueAtFieldPointer(cluster *v1alpha1.Cluster, fieldPtr any, ref *v1alpha1.Cluster, value any) {
	// Get the field path to determine where to set the value
	path := getFieldPathFromPointer(fieldPtr, ref)
	if path == "" {
		return
	}
	
	// Use the existing getFieldByPath function to navigate to the field
	targetFieldPtr := getFieldByPath(cluster, path)
	if targetFieldPtr == nil {
		return
	}
	
	// Set the value using reflection
	fieldVal := reflect.ValueOf(targetFieldPtr)
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
