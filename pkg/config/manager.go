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
	viper   *viper.Viper
	cluster *v1alpha1.Cluster
}

// NewManager creates a new configuration manager.
func NewManager() *Manager {
	v := initializeViper()
	
	// Set all defaults using config defaults
	setViperDefaultsFromConfigDefaults(v)
	
	return &Manager{
		viper:   v,
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

	// Apply all configuration and defaults through the config defaults
	m.setClusterFromConfigDefaults(cluster)

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

// setViperDefaultsFromConfigDefaults sets all configuration defaults in Viper using config defaults.
func setViperDefaultsFromConfigDefaults(v *viper.Viper) {
	configDefaults := GetConfigDefaults()

	for _, configDefault := range configDefaults {
		// Get the path dynamically from the field selector
		path := configDefault.GetPath()
		
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

// setClusterFromConfigDefaults applies all configuration values to the cluster using config defaults.
func (m *Manager) setClusterFromConfigDefaults(cluster *v1alpha1.Cluster) {
	configDefaults := GetConfigDefaults()

	for _, configDefault := range configDefaults {
		// Get the path dynamically from the field selector
		path := configDefault.GetPath()
		
		// Get value from Viper and set it in the cluster
		value := m.getTypedValueFromViperByPath(path)
		configDefault.SetValue(cluster, value)
	}
}

// getTypedValueFromViperByPath retrieves a properly typed value from Viper based on the path.
// This now uses dynamic type inference instead of hardcoded paths.
func (m *Manager) getTypedValueFromViperByPath(path string) any {
	// Find the corresponding ConfigDefault to determine the expected type
	configDefaults := GetConfigDefaults()
	
	for _, configDefault := range configDefaults {
		if configDefault.GetPath() == path {
			// Get the expected type from the field selector
			dummy := &v1alpha1.Cluster{}
			fieldPtr := configDefault.FieldSelector(dummy)
			
			if fieldPtr == nil {
				continue
			}
			
			fieldVal := reflect.ValueOf(fieldPtr)
			if fieldVal.Kind() != reflect.Ptr {
				continue
			}
			
			fieldType := fieldVal.Elem().Type()
			
			// Get the value from Viper based on the field type
			rawValue := m.getValueFromViperByType(path, fieldType)
			
			// Convert the value to the appropriate type
			return convertValueToFieldType(rawValue, fieldType)
		}
	}
	
	// Fallback to string for unknown paths
	return m.viper.GetString(path)
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
