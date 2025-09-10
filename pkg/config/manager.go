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
	// Set defaults in Viper if not already set
	m.viper.SetDefault("metadata.name", "ksail-default")

	// Let Viper handle precedence automatically: flags > env vars > config file > defaults
	cluster.Metadata.Name = m.viper.GetString("metadata.name")
}

// setSpecFromConfig sets spec values from configuration with defaults.
func (m *Manager) setSpecFromConfig(cluster *v1alpha1.Cluster) {
	// Set defaults in Viper if not already set
	m.viper.SetDefault("spec.distributionconfig", "kind.yaml")
	m.viper.SetDefault("spec.sourcedirectory", "k8s")
	m.viper.SetDefault("spec.distribution", string(v1alpha1.DistributionKind))
	m.viper.SetDefault("spec.reconciliationtool", string(v1alpha1.ReconciliationToolKubectl))
	m.viper.SetDefault("spec.cni", string(v1alpha1.CNIDefault))
	m.viper.SetDefault("spec.csi", string(v1alpha1.CSIDefault))
	m.viper.SetDefault("spec.ingresscontroller", string(v1alpha1.IngressControllerDefault))
	m.viper.SetDefault("spec.gatewaycontroller", string(v1alpha1.GatewayControllerDefault))

	// Let Viper handle precedence automatically: flags > env vars > config file > defaults
	cluster.Spec.DistributionConfig = m.viper.GetString("spec.distributionconfig")
	cluster.Spec.SourceDirectory = m.viper.GetString("spec.sourcedirectory")

	// Distribution
	if distStr := m.viper.GetString("spec.distribution"); distStr != "" {
		var distribution v1alpha1.Distribution
		err := distribution.Set(distStr)
		if err == nil {
			cluster.Spec.Distribution = distribution
		} else {
			cluster.Spec.Distribution = v1alpha1.DistributionKind
		}
	}

	// Reconciliation Tool
	if tool := m.viper.GetString("spec.reconciliationtool"); tool != "" {
		var reconciliationTool v1alpha1.ReconciliationTool
		err := reconciliationTool.Set(tool)
		if err == nil {
			cluster.Spec.ReconciliationTool = reconciliationTool
		} else {
			cluster.Spec.ReconciliationTool = v1alpha1.ReconciliationToolKubectl
		}
	}

	// Other fields use simple string assignment
	cluster.Spec.CNI = v1alpha1.CNI(m.viper.GetString("spec.cni"))
	cluster.Spec.CSI = v1alpha1.CSI(m.viper.GetString("spec.csi"))
	cluster.Spec.IngressController = v1alpha1.IngressController(m.viper.GetString("spec.ingresscontroller"))
	cluster.Spec.GatewayController = v1alpha1.GatewayController(m.viper.GetString("spec.gatewaycontroller"))
}

const defaultConnectionTimeoutMinutes = 5

// setConnectionFromConfig sets connection values from configuration with defaults.
func (m *Manager) setConnectionFromConfig(cluster *v1alpha1.Cluster) {
	// Set defaults in Viper if not already set
	m.viper.SetDefault("spec.connection.kubeconfig", "~/.kube/config")
	m.viper.SetDefault("spec.connection.context", "kind-ksail-default")
	m.viper.SetDefault("spec.connection.timeout", "5m")

	// Let Viper handle precedence automatically: flags > env vars > config file > defaults
	cluster.Spec.Connection.Kubeconfig = m.viper.GetString("spec.connection.kubeconfig")
	cluster.Spec.Connection.Context = m.viper.GetString("spec.connection.context")

	// Timeout requires parsing
	if timeoutStr := m.viper.GetString("spec.connection.timeout"); timeoutStr != "" {
		if duration, err := time.ParseDuration(timeoutStr); err == nil {
			cluster.Spec.Connection.Timeout = metav1.Duration{Duration: duration}
		} else {
			cluster.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
		}
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
