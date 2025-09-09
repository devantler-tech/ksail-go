// Package v1alpha1 provides model definitions for a KSail cluster.
package v1alpha1

import (
	"errors"
	"fmt"
	"strings"
	"time"

	k8sutils "github.com/devantler-tech/ksail-go/internal/utils/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// --- Errors ---

// ErrInvalidDistribution is returned when an invalid distribution is specified.
var ErrInvalidDistribution = errors.New("invalid distribution")

// ErrInvalidReconciliationTool is returned when an invalid reconciliation tool is specified.
var ErrInvalidReconciliationTool = errors.New("invalid reconciliation tool")

// ErrInvalidContainerEngine is returned when an invalid container engine is specified.
var ErrInvalidContainerEngine = errors.New("invalid container engine")

// --- Constructors ---

// CreateDefaultMetadata creates a default metav1.ObjectMeta with the given name.
func CreateDefaultMetadata(name string) metav1.ObjectMeta {
	metadata := k8sutils.NewEmptyObjectMeta()
	metadata.Name = name
	metadata.OwnerReferences = []metav1.OwnerReference{}
	metadata.Finalizers = []string{}
	metadata.ManagedFields = []metav1.ManagedFieldsEntry{}

	return metadata
}

// NewCluster creates a new KSail cluster with the given options.
func NewCluster(options ...func(*Cluster)) *Cluster {
	cluster := &Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       Kind,
			APIVersion: APIVersion,
		},
		Metadata: k8sutils.NewEmptyObjectMeta(),
		Spec: Spec{
			Connection: Connection{
				Kubeconfig: "",
				Context:    "",
				Timeout:    metav1.Duration{Duration: time.Duration(0)},
			},
			DistributionConfig: "",
			SourceDirectory:    "",
			Distribution:       "",
			CNI:                "",
			CSI:                "",
			IngressController:  "",
			GatewayController:  "",
			ReconciliationTool: "",
			Options: Options{
				Kind:      OptionsKind{},
				K3d:       OptionsK3d{},
				Tind:      OptionsTind{},
				EKS:       OptionsEKS{AWSProfile: ""},
				Cilium:    OptionsCilium{},
				Kubectl:   OptionsKubectl{},
				Flux:      OptionsFlux{},
				ArgoCD:    OptionsArgoCD{},
				Helm:      OptionsHelm{},
				Kustomize: OptionsKustomize{},
			},
		},
	}
	for _, opt := range options {
		opt(cluster)
	}

	cluster.SetDefaults()

	return cluster
}

// WithMetadataName sets the name of the cluster.
func WithMetadataName(name string) func(*Cluster) {
	return func(c *Cluster) {
		c.Metadata.Name = name
	}
}

// WithSpecDistribution sets the distribution of the cluster.
func WithSpecDistribution(distribution Distribution) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.Distribution = distribution
	}
}

// WithSpecConnectionKubeconfig sets the kubeconfig for the cluster.
func WithSpecConnectionKubeconfig(kubeconfig string) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.Connection.Kubeconfig = kubeconfig
	}
}

// WithSpecConnectionContext sets the context for the cluster.
func WithSpecConnectionContext(context string) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.Connection.Context = context
	}
}

// WithSpecConnectionTimeout sets the timeout for the cluster.
func WithSpecConnectionTimeout(timeout metav1.Duration) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.Connection.Timeout = timeout
	}
}

// WithSpecCNI sets the CNI for the cluster.
func WithSpecCNI(cni CNI) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.CNI = cni
	}
}

// WithSpecCSI sets the CSI implementation on the cluster spec.
func WithSpecCSI(csi CSI) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.CSI = csi
	}
}

// WithSpecIngressController sets the ingress controller on the cluster spec.
func WithSpecIngressController(ingressController IngressController) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.IngressController = ingressController
	}
}

// WithSpecGatewayController sets the gateway controller on the cluster spec.
func WithSpecGatewayController(gatewayController GatewayController) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.GatewayController = gatewayController
	}
}

// WithSpecReconciliationTool sets the deployment tool on the cluster spec.
func WithSpecReconciliationTool(reconciliationTool ReconciliationTool) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.ReconciliationTool = reconciliationTool
	}
}

// --- Defaults ---

// SetDefaults sets default values for the Cluster fields if they are not already set.
func (c *Cluster) SetDefaults() {
	c.setMetadataDefaults()
	c.setSpecDefaults()
	c.setSpecConnectionDefaults()
}

// ConfigSource provides a simple interface for configuration values.
type ConfigSource interface {
	GetString(key string) string
}

// SetDefaultsFromConfigSource sets default values for the Cluster fields using a configuration source.
// This method allows integration with Viper or any other configuration system.
func (c *Cluster) SetDefaultsFromConfigSource(configSource ConfigSource) {
	c.setMetadataDefaultsFromConfigSource(configSource)
	c.setSpecDefaultsFromConfigSource(configSource)
	c.setSpecConnectionDefaultsFromConfigSource(configSource)
}

func (c *Cluster) setMetadataDefaults() {
	if c.Metadata.Name == "" {
		c.Metadata.Name = "ksail-default"
	}
}

func (c *Cluster) setSpecDefaults() {
	if c.Spec.DistributionConfig == "" {
		c.Spec.DistributionConfig = "kind.yaml"
	}

	if c.Spec.SourceDirectory == "" {
		c.Spec.SourceDirectory = "k8s"
	}

	if c.Spec.Distribution == "" {
		c.Spec.Distribution = DistributionKind
	}

	if c.Spec.ReconciliationTool == "" {
		c.Spec.ReconciliationTool = ReconciliationToolKubectl
	}

	if c.Spec.CNI == "" {
		c.Spec.CNI = CNIDefault
	}

	if c.Spec.CSI == "" {
		c.Spec.CSI = CSIDefault
	}

	if c.Spec.IngressController == "" {
		c.Spec.IngressController = IngressControllerDefault
	}

	if c.Spec.GatewayController == "" {
		c.Spec.GatewayController = GatewayControllerDefault
	}
}

const defaultConnectionTimeoutMinutes = 5

func (c *Cluster) setSpecConnectionDefaults() {
	if c.Spec.Connection.Kubeconfig == "" {
		c.Spec.Connection.Kubeconfig = "~/.kube/config"
	}

	if c.Spec.Connection.Context == "" {
		c.Spec.Connection.Context = "kind-ksail-default"
	}

	if c.Spec.Connection.Timeout.Duration == 0 {
		c.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
	}
}

func (c *Cluster) setMetadataDefaultsFromConfigSource(configSource ConfigSource) {
	if c.Metadata.Name == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if name := configSource.GetString("metadata.name"); name != "" {
			c.Metadata.Name = name
		} else if name := configSource.GetString("cluster.name"); name != "" {
			// Backward compatibility for flat config structure
			c.Metadata.Name = name
		} else {
			c.Metadata.Name = "ksail-default"
		}
	}
}

func (c *Cluster) setSpecDefaultsFromConfigSource(configSource ConfigSource) {
	if c.Spec.DistributionConfig == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if distConfig := configSource.GetString("spec.distributionConfig"); distConfig != "" {
			c.Spec.DistributionConfig = distConfig
		} else {
			c.Spec.DistributionConfig = "kind.yaml"
		}
	}

	if c.Spec.SourceDirectory == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if sourceDir := configSource.GetString("spec.sourceDirectory"); sourceDir != "" {
			c.Spec.SourceDirectory = sourceDir
		} else {
			c.Spec.SourceDirectory = "k8s"
		}
	}

	if c.Spec.Distribution == "" {
		// Try CLI flag first (highest precedence), then hierarchical structure (ksail.yaml format)
		var distStr string
		if distStr = configSource.GetString("distribution"); distStr == "Kind" {
			// If CLI is still default, try config file
			if fileDistStr := configSource.GetString("spec.distribution"); fileDistStr != "" {
				distStr = fileDistStr
			}
		}
		
		if distStr != "" {
			var distribution Distribution
			if err := distribution.Set(distStr); err == nil {
				c.Spec.Distribution = distribution
			} else {
				c.Spec.Distribution = DistributionKind
			}
		} else {
			c.Spec.Distribution = DistributionKind
		}
	}

	if c.Spec.ReconciliationTool == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if tool := configSource.GetString("spec.reconciliationTool"); tool != "" {
			var reconciliationTool ReconciliationTool
			if err := reconciliationTool.Set(tool); err == nil {
				c.Spec.ReconciliationTool = reconciliationTool
			} else {
				c.Spec.ReconciliationTool = ReconciliationToolKubectl
			}
		} else {
			c.Spec.ReconciliationTool = ReconciliationToolKubectl
		}
	}

	if c.Spec.CNI == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if cni := configSource.GetString("spec.cni"); cni != "" {
			c.Spec.CNI = CNI(cni)
		} else {
			c.Spec.CNI = CNIDefault
		}
	}

	if c.Spec.CSI == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if csi := configSource.GetString("spec.csi"); csi != "" {
			c.Spec.CSI = CSI(csi)
		} else {
			c.Spec.CSI = CSIDefault
		}
	}

	if c.Spec.IngressController == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if ingress := configSource.GetString("spec.ingressController"); ingress != "" {
			c.Spec.IngressController = IngressController(ingress)
		} else {
			c.Spec.IngressController = IngressControllerDefault
		}
	}

	if c.Spec.GatewayController == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if gateway := configSource.GetString("spec.gatewayController"); gateway != "" {
			c.Spec.GatewayController = GatewayController(gateway)
		} else {
			c.Spec.GatewayController = GatewayControllerDefault
		}
	}
}

func (c *Cluster) setSpecConnectionDefaultsFromConfigSource(configSource ConfigSource) {
	if c.Spec.Connection.Kubeconfig == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if kubeconfig := configSource.GetString("spec.connection.kubeconfig"); kubeconfig != "" {
			c.Spec.Connection.Kubeconfig = kubeconfig
		} else if kubeconfig := configSource.GetString("cluster.connection.kubeconfig"); kubeconfig != "" {
			// Backward compatibility for flat config structure
			c.Spec.Connection.Kubeconfig = kubeconfig
		} else {
			c.Spec.Connection.Kubeconfig = "~/.kube/config"
		}
	}

	if c.Spec.Connection.Context == "" {
		// Try hierarchical structure first (ksail.yaml format)
		if context := configSource.GetString("spec.connection.context"); context != "" {
			c.Spec.Connection.Context = context
		} else if context := configSource.GetString("cluster.connection.context"); context != "" {
			// Backward compatibility for flat config structure
			c.Spec.Connection.Context = context
		} else {
			c.Spec.Connection.Context = "kind-ksail-default"
		}
	}

	if c.Spec.Connection.Timeout.Duration == 0 {
		// Try hierarchical structure first (ksail.yaml format)
		if timeoutStr := configSource.GetString("spec.connection.timeout"); timeoutStr != "" {
			if timeout, err := time.ParseDuration(timeoutStr); err == nil {
				c.Spec.Connection.Timeout = metav1.Duration{Duration: timeout}
			} else {
				c.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
			}
		} else if timeoutStr := configSource.GetString("cluster.connection.timeout"); timeoutStr != "" {
			// Backward compatibility for flat config structure
			if timeout, err := time.ParseDuration(timeoutStr); err == nil {
				c.Spec.Connection.Timeout = metav1.Duration{Duration: timeout}
			} else {
				c.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
			}
		} else {
			c.Spec.Connection.Timeout = metav1.Duration{Duration: time.Duration(defaultConnectionTimeoutMinutes) * time.Minute}
		}
	}
}

// --- Getters and Setters ---

// Set for Distribution.
func (d *Distribution) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, dist := range validDistributions() {
		if strings.EqualFold(value, string(dist)) {
			*d = dist

			return nil
		}
	}

	return fmt.Errorf("%w: %s (valid options: %s, %s, %s)",
		ErrInvalidDistribution, value, DistributionKind, DistributionK3d, DistributionTind)
}

// Set for ReconciliationTool.
func (d *ReconciliationTool) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, tool := range validReconciliationTools() {
		if strings.EqualFold(value, string(tool)) {
			*d = tool

			return nil
		}
	}

	return fmt.Errorf("%w: %s (valid options: %s, %s, %s)",
		ErrInvalidReconciliationTool, value, ReconciliationToolKubectl, ReconciliationToolFlux, ReconciliationToolArgoCD)
}

// -- pflags --

// String returns the string representation of the Distribution.
func (d *Distribution) String() string {
	return string(*d)
}

// Type returns the type of the Distribution.
func (d *Distribution) Type() string {
	return "Distribution"
}

// String returns the string representation of the ReconciliationTool.
func (d *ReconciliationTool) String() string {
	return string(*d)
}

// Type returns the type of the ReconciliationTool.
func (d *ReconciliationTool) Type() string {
	return "ReconciliationTool"
}
