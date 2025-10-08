package v1alpha1

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// --- Errors ---

// ErrInvalidDistribution is returned when an invalid distribution is specified.
var ErrInvalidDistribution = errors.New("invalid distribution")

// ErrInvalidGitOpsEngine is returned when an invalid GitOps engine is specified.
var ErrInvalidGitOpsEngine = errors.New("invalid GitOps engine")

// ErrInvalidCNI is returned when an invalid CNI is specified.
var ErrInvalidCNI = errors.New("invalid CNI")

// ErrInvalidCSI is returned when an invalid CSI is specified.
var ErrInvalidCSI = errors.New("invalid CSI")

// ErrInvalidIngressController is returned when an invalid ingress controller is specified.
var ErrInvalidIngressController = errors.New("invalid ingress controller")

// ErrInvalidGatewayController is returned when an invalid gateway controller is specified.
var ErrInvalidGatewayController = errors.New("invalid gateway controller")

const (
	// Group is the API group for KSail.
	Group = "ksail.dev"
	// Version is the API version for KSail.
	Version = "v1alpha1"
	// Kind is the kind for KSail clusters.
	Kind = "Cluster"
	// APIVersion is the full API version for KSail.
	APIVersion = Group + "/" + Version
)

// Cluster represents a KSail cluster configuration including API metadata and desired state.
// It contains TypeMeta for API versioning information and Spec for the cluster specification.
type Cluster struct {
	metav1.TypeMeta `json:",inline"`

	Spec Spec `json:"spec,omitzero"`
}

// Spec defines the desired state of a KSail cluster.
type Spec struct {
	DistributionConfig string            `json:"distributionConfig,omitzero"`
	SourceDirectory    string            `json:"sourceDirectory,omitzero"`
	Connection         Connection        `json:"connection,omitzero"`
	Distribution       Distribution      `json:"distribution,omitzero"`
	CNI                CNI               `json:"cni,omitzero"`
	CSI                CSI               `json:"csi,omitzero"`
	IngressController  IngressController `json:"ingressController,omitzero"`
	GatewayController  GatewayController `json:"gatewayController,omitzero"`
	GitOpsEngine       GitOpsEngine      `json:"gitOpsEngine,omitzero"`
	Options            Options           `json:"options,omitzero"`
}

// Connection defines connection options for a KSail cluster.
type Connection struct {
	Kubeconfig string          `json:"kubeconfig,omitzero"`
	Context    string          `json:"context,omitzero"`
	Timeout    metav1.Duration `json:"timeout,omitzero"`
}

// Distribution defines the distribution options for a KSail cluster.
type Distribution string

const (
	// DistributionKind is the kind distribution.
	DistributionKind Distribution = "Kind"
	// DistributionK3d is the K3d distribution.
	DistributionK3d Distribution = "K3d"
)

// validDistributions returns supported distribution values.
func validDistributions() []Distribution {
	return []Distribution{DistributionK3d, DistributionKind}
}

// validCNIs returns supported CNI values.
func validCNIs() []CNI {
	return []CNI{CNIDefault, CNICilium}
}

// validCSIs returns supported CSI values.
func validCSIs() []CSI {
	return []CSI{CSIDefault, CSILocalPathStorage}
}

// validIngressControllers returns supported ingress controller values.
func validIngressControllers() []IngressController {
	return []IngressController{
		IngressControllerDefault,
		IngressControllerTraefik,
		IngressControllerNone,
	}
}

// validGatewayControllers returns supported gateway controller values.
func validGatewayControllers() []GatewayController {
	return []GatewayController{
		GatewayControllerDefault,
		GatewayControllerTraefik,
		GatewayControllerCilium,
		GatewayControllerNone,
	}
}

// CNI defines the CNI options for a KSail cluster.
type CNI string

const (
	// CNIDefault is the default CNI.
	CNIDefault CNI = "Default"
	// CNICilium is the Cilium CNI.
	CNICilium CNI = "Cilium"
)

// CSI defines the CSI options for a KSail cluster.
type CSI string

const (
	// CSIDefault is the default CSI.
	CSIDefault CSI = "Default"
	// CSILocalPathStorage is the LocalPathStorage CSI.
	CSILocalPathStorage CSI = "LocalPathStorage"
)

// IngressController defines the Ingress Controller options for a KSail cluster.
type IngressController string

const (
	// IngressControllerDefault is the default Ingress Controller.
	IngressControllerDefault IngressController = "Default"
	// IngressControllerTraefik is the Traefik Ingress Controller.
	IngressControllerTraefik IngressController = "Traefik"
	// IngressControllerNone is no Ingress Controller.
	IngressControllerNone IngressController = "None"
)

// GatewayController defines the Gateway Controller options for a KSail cluster.
type GatewayController string

const (
	// GatewayControllerDefault is the default Gateway Controller.
	GatewayControllerDefault GatewayController = "Default"
	// GatewayControllerTraefik is the Traefik Gateway Controller.
	GatewayControllerTraefik GatewayController = "Traefik"
	// GatewayControllerCilium is the Cilium Gateway Controller.
	GatewayControllerCilium GatewayController = "Cilium"
	// GatewayControllerNone is no Gateway Controller.
	GatewayControllerNone GatewayController = "None"
)

// GitOpsEngine defines the GitOps Engine options for a KSail cluster.
type GitOpsEngine string

const (
	// GitOpsEngineNone is no GitOps engine.
	GitOpsEngineNone GitOpsEngine = "None"
)

// validGitOpsEngines enumerates supported GitOps engine values.
func validGitOpsEngines() []GitOpsEngine {
	return []GitOpsEngine{
		GitOpsEngineNone,
	}
}

// Options holds optional settings for distributions, networking, and deployment tools.
type Options struct {
	Kind OptionsKind `json:"kind,omitzero"`
	K3d  OptionsK3d  `json:"k3d,omitzero"`

	Cilium OptionsCilium `json:"cilium,omitzero"`

	Flux   OptionsFlux   `json:"flux,omitzero"`
	ArgoCD OptionsArgoCD `json:"argocd,omitzero"`

	Helm      OptionsHelm      `json:"helm,omitzero"`
	Kustomize OptionsKustomize `json:"kustomize,omitzero"`
}

// OptionsKind defines options specific to the Kind distribution.
type OptionsKind struct {
	// Add any specific fields for the Kind distribution here.
}

// OptionsK3d defines options specific to the K3d distribution.
type OptionsK3d struct {
	// Add any specific fields for the K3d distribution here.
}

// OptionsCilium defines options for the Cilium CNI.
type OptionsCilium struct {
	// Add any specific fields for the Cilium CNI here.
}

// OptionsFlux defines options for the Flux deployment tool.
type OptionsFlux struct {
	// Add any specific fields for the Flux distribution here.
}

// OptionsArgoCD defines options for the ArgoCD deployment tool.
type OptionsArgoCD struct {
	// Add any specific fields for the ArgoCD distribution here.
}

// OptionsHelm defines options for the Helm tool.
type OptionsHelm struct {
	// Add any specific fields for the Helm distribution here.
}

// OptionsKustomize defines options for the Kustomize tool.
type OptionsKustomize struct {
	// Add any specific fields for the Kustomize distribution here.
}

// --- Setters for pflags ---

// Set for Distribution.
func (d *Distribution) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, dist := range validDistributions() {
		if strings.EqualFold(value, string(dist)) {
			*d = dist

			return nil
		}
	}

	return fmt.Errorf("%w: %s (valid options: %s, %s)",
		ErrInvalidDistribution, value, DistributionKind, DistributionK3d)
}

// Set for GitOpsEngine.
func (d *GitOpsEngine) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, tool := range validGitOpsEngines() {
		if strings.EqualFold(value, string(tool)) {
			*d = tool

			return nil
		}
	}

	return fmt.Errorf(
		"%w: %s (valid options: %s)",
		ErrInvalidGitOpsEngine,
		value,
		GitOpsEngineNone,
	)
}

// Set for CNI.
func (c *CNI) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, cni := range validCNIs() {
		if strings.EqualFold(value, string(cni)) {
			*c = cni

			return nil
		}
	}

	return fmt.Errorf("%w: %s (valid options: %s, %s)",
		ErrInvalidCNI, value, CNIDefault, CNICilium)
}

// Set for CSI.
func (c *CSI) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, csi := range validCSIs() {
		if strings.EqualFold(value, string(csi)) {
			*c = csi

			return nil
		}
	}

	return fmt.Errorf("%w: %s (valid options: %s, %s)",
		ErrInvalidCSI, value, CSIDefault, CSILocalPathStorage)
}

// Set for IngressController.
func (i *IngressController) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, ic := range validIngressControllers() {
		if strings.EqualFold(value, string(ic)) {
			*i = ic

			return nil
		}
	}

	return fmt.Errorf(
		"%w: %s (valid options: %s, %s, %s)",
		ErrInvalidIngressController,
		value,
		IngressControllerDefault,
		IngressControllerTraefik,
		IngressControllerNone,
	)
}

// Set for GatewayController.
func (g *GatewayController) Set(value string) error {
	// Check against constant values with case-insensitive comparison
	for _, gc := range validGatewayControllers() {
		if strings.EqualFold(value, string(gc)) {
			*g = gc

			return nil
		}
	}

	return fmt.Errorf(
		"%w: %s (valid options: %s, %s, %s, %s)",
		ErrInvalidGatewayController,
		value,
		GatewayControllerDefault,
		GatewayControllerTraefik,
		GatewayControllerCilium,
		GatewayControllerNone,
	)
}

// IsValid checks if the distribution value is supported.
func (d *Distribution) IsValid() bool {
	return slices.Contains(validDistributions(), *d)
}

// String returns the string representation of the Distribution.
func (d *Distribution) String() string {
	return string(*d)
}

// Type returns the type of the Distribution.
func (d *Distribution) Type() string {
	return "Distribution"
}

// String returns the string representation of the GitOpsEngine.
func (d *GitOpsEngine) String() string {
	return string(*d)
}

// Type returns the type of the GitOpsEngine.
func (d *GitOpsEngine) Type() string {
	return "GitOpsEngine"
}

// String returns the string representation of the CNI.
func (c *CNI) String() string {
	return string(*c)
}

// Type returns the type of the CNI.
func (c *CNI) Type() string {
	return "CNI"
}

// String returns the string representation of the CSI.
func (c *CSI) String() string {
	return string(*c)
}

// Type returns the type of the CSI.
func (c *CSI) Type() string {
	return "CSI"
}

// String returns the string representation of the IngressController.
func (i *IngressController) String() string {
	return string(*i)
}

// Type returns the type of the IngressController.
func (i *IngressController) Type() string {
	return "IngressController"
}

// String returns the string representation of the GatewayController.
func (g *GatewayController) String() string {
	return string(*g)
}

// Type returns the type of the GatewayController.
func (g *GatewayController) Type() string {
	return "GatewayController"
}
