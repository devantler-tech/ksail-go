// Package cluster provides model definitions for a KSail cluster.
package v1alpha1

import (
	"errors"
	"fmt"
	"strings"
	"time"

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

// NewCluster creates a new KSail cluster with the given options.
func NewCluster(options ...func(*Cluster)) *Cluster {
	cluster := &Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       Kind,
			APIVersion: APIVersion,
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

// WithSpecContainerEngine sets the container engine of the cluster.
func WithSpecContainerEngine(engine ContainerEngine) func(*Cluster) {
	return func(c *Cluster) {
		c.Spec.ContainerEngine = engine
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

	if c.Spec.ContainerEngine == "" {
		c.Spec.ContainerEngine = ContainerEngineDocker
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

// Set for ContainerEngine.
func (e *ContainerEngine) Set(value string) error {
	for _, engine := range validContainerEngines() {
		if strings.EqualFold(value, string(engine)) {
			*e = engine

			return nil
		}
	}

	return fmt.Errorf(
		"%w: %s (valid options: %s, %s)",
		ErrInvalidContainerEngine,
		value,
		ContainerEngineDocker,
		ContainerEnginePodman,
	)
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

// String returns the string representation of the ContainerEngine.
func (e *ContainerEngine) String() string { return string(*e) }

// Type returns the type of the ContainerEngine.
func (e *ContainerEngine) Type() string { return "ContainerEngine" }
