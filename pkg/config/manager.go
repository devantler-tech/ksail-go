package config

import (
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

		v.SetDefault(configDefault.Path, viperValue)
	}
}

// setClusterFromConfigDefaults applies all configuration values to the cluster using config defaults.
func (m *Manager) setClusterFromConfigDefaults(cluster *v1alpha1.Cluster) {
	configDefaults := GetConfigDefaults()

	for _, configDefault := range configDefaults {
		// Get value from Viper and set it in the cluster
		value := m.getTypedValueFromViperByPath(configDefault.Path)
		configDefault.SetValue(cluster, value)
	}
}

// getTypedValueFromViperByPath retrieves a properly typed value from Viper based on the path.
func (m *Manager) getTypedValueFromViperByPath(path string) any {
	// Determine expected type based on path
	switch path {
	case "metadata.name", "spec.distributionconfig", "spec.sourcedirectory",
		"spec.connection.kubeconfig", "spec.connection.context":
		return m.viper.GetString(path)
	case "spec.distribution":
		distStr := m.viper.GetString(path)
		var distribution v1alpha1.Distribution
		if err := distribution.Set(distStr); err == nil {
			return distribution
		}
		return v1alpha1.DistributionKind
	case "spec.reconciliationtool":
		toolStr := m.viper.GetString(path)
		var tool v1alpha1.ReconciliationTool
		if err := tool.Set(toolStr); err == nil {
			return tool
		}
		return v1alpha1.ReconciliationToolKubectl
	case "spec.cni":
		return v1alpha1.CNI(m.viper.GetString(path))
	case "spec.csi":
		return v1alpha1.CSI(m.viper.GetString(path))
	case "spec.ingresscontroller":
		return v1alpha1.IngressController(m.viper.GetString(path))
	case "spec.gatewaycontroller":
		return v1alpha1.GatewayController(m.viper.GetString(path))
	case "spec.connection.timeout":
		timeoutStr := m.viper.GetString(path)
		if duration, err := time.ParseDuration(timeoutStr); err == nil {
			return metav1.Duration{Duration: duration}
		}
		return metav1.Duration{Duration: 5 * time.Minute}
	default:
		// Fallback to string for unknown paths
		return m.viper.GetString(path)
	}
}
