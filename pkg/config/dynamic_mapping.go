package config

import (
	"reflect"
	"strings"
	"time"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultValueMapping represents default values for configuration fields.
type DefaultValueMapping struct {
	Path         string
	DefaultValue any
}

// getDefaultValueMappings provides default values for all fields in v1alpha1.Cluster.
func getDefaultValueMappings() []DefaultValueMapping {
	return []DefaultValueMapping{
		// Metadata defaults
		{"metadata.name", "ksail-default"},

		// Spec defaults
		{"spec.distributionconfig", "kind.yaml"},
		{"spec.sourcedirectory", "k8s"},
		{"spec.distribution", v1alpha1.DistributionKind},
		{"spec.reconciliationtool", v1alpha1.ReconciliationToolKubectl},
		{"spec.cni", v1alpha1.CNIDefault},
		{"spec.csi", v1alpha1.CSIDefault},
		{"spec.ingresscontroller", v1alpha1.IngressControllerDefault},
		{"spec.gatewaycontroller", v1alpha1.GatewayControllerDefault},

		// Connection defaults
		{"spec.connection.kubeconfig", "~/.kube/config"},
		{"spec.connection.context", "kind-ksail-default"},
		{"spec.connection.timeout", "5m"},
	}
}

// setViperDefaultsDynamic sets all configuration defaults in Viper using field mappings.
func setViperDefaultsDynamic(v *viper.Viper) {
	mappings := getDefaultValueMappings()

	for _, mapping := range mappings {
		// Convert typed default value to appropriate format for Viper
		var viperValue any
		switch val := mapping.DefaultValue.(type) {
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

		v.SetDefault(mapping.Path, viperValue)
	}
}

// setClusterFromConfigDynamic applies all configuration values to the cluster using reflection.
func (m *Manager) setClusterFromConfigDynamic(cluster *v1alpha1.Cluster) {
	mappings := getDefaultValueMappings()

	for _, mapping := range mappings {
		value := m.getTypedValueFromViper(mapping.Path)
		m.setFieldValueByPath(cluster, mapping.Path, value)
	}
}

// getTypedValueFromViper retrieves a properly typed value from Viper based on the path and expected type.
func (m *Manager) getTypedValueFromViper(path string) any {
	// Determine expected type based on path and get appropriate value
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

// setFieldValueByPath uses reflection to set a field value by its dot-separated path.
func (m *Manager) setFieldValueByPath(cluster *v1alpha1.Cluster, path string, value any) {
	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return
	}

	clusterVal := reflect.ValueOf(cluster).Elem()
	field := clusterVal

	// Navigate to the field using the path
	for i, part := range parts {
		if !field.IsValid() || field.Kind() != reflect.Struct {
			return
		}

		// Convert camelCase field names to proper struct field names
		fieldName := ""
		switch part {
		case "metadata":
			fieldName = "Metadata"
		case "spec":
			fieldName = "Spec"
		case "name":
			fieldName = "Name"
		case "distributionconfig":
			fieldName = "DistributionConfig"
		case "sourcedirectory":
			fieldName = "SourceDirectory"
		case "distribution":
			fieldName = "Distribution"
		case "reconciliationtool":
			fieldName = "ReconciliationTool"
		case "cni":
			fieldName = "CNI"
		case "csi":
			fieldName = "CSI"
		case "ingresscontroller":
			fieldName = "IngressController"
		case "gatewaycontroller":
			fieldName = "GatewayController"
		case "connection":
			fieldName = "Connection"
		case "kubeconfig":
			fieldName = "Kubeconfig"
		case "context":
			fieldName = "Context"
		case "timeout":
			fieldName = "Timeout"
		default:
			return // Unknown field name
		}

		field = field.FieldByName(fieldName)
		if !field.IsValid() {
			return
		}

		// If this is the last part, set the value
		if i == len(parts)-1 {
			if field.CanSet() && reflect.TypeOf(value) == field.Type() {
				field.Set(reflect.ValueOf(value))
			}
		}
	}
}