package config

import (
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/internal/utils/k8s"
	"github.com/spf13/pflag"
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

// BindPFlag binds a CLI flag to a configuration key.
func (m *Manager) BindPFlag(key string, flag *pflag.Flag) error {
	return m.viper.BindPFlag(key, flag)
}