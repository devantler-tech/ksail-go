package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewCluster creates a new Cluster instance with minimal required structure.
// All default values are now handled by the configuration system via field selectors.
func NewCluster() *Cluster {
	return &Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       Kind,
			APIVersion: APIVersion,
		},
		Spec: NewClusterSpec(),
	}
}

// NewClusterSpec creates a new Spec with default values.
func NewClusterSpec() Spec {
	return Spec{
		DistributionConfig: "",
		SourceDirectory:    "",
		Connection:         NewClusterConnection(),
		Distribution:       "",
		CNI:                "",
		CSI:                "",
		IngressController:  "",
		GatewayController:  "",
		ReconciliationTool: "",
		Options:            NewClusterOptions(),
	}
}

// NewClusterConnection creates a new Connection with default values.
func NewClusterConnection() Connection {
	return Connection{
		Kubeconfig: "",
		Context:    "",
		Timeout:    metav1.Duration{Duration: 0},
	}
}

// NewClusterOptions creates a new Options with default values.
func NewClusterOptions() Options {
	return Options{
		Kind:      NewClusterOptionsKind(),
		K3d:       NewClusterOptionsK3d(),
		Tind:      NewClusterOptionsTind(),
		Cilium:    NewClusterOptionsCilium(),
		Kubectl:   NewClusterOptionsKubectl(),
		Flux:      NewClusterOptionsFlux(),
		ArgoCD:    NewClusterOptionsArgoCD(),
		Helm:      NewClusterOptionsHelm(),
		Kustomize: NewClusterOptionsKustomize(),
	}
}

// NewClusterOptionsKind creates a new OptionsKind with default values.
func NewClusterOptionsKind() OptionsKind {
	return OptionsKind{}
}

// NewClusterOptionsK3d creates a new OptionsK3d with default values.
func NewClusterOptionsK3d() OptionsK3d {
	return OptionsK3d{}
}

// NewClusterOptionsTind creates a new OptionsTind with default values.
func NewClusterOptionsTind() OptionsTind {
	return OptionsTind{}
}

// NewClusterOptionsCilium creates a new OptionsCilium with default values.
func NewClusterOptionsCilium() OptionsCilium {
	return OptionsCilium{}
}

// NewClusterOptionsKubectl creates a new OptionsKubectl with default values.
func NewClusterOptionsKubectl() OptionsKubectl {
	return OptionsKubectl{}
}

// NewClusterOptionsFlux creates a new OptionsFlux with default values.
func NewClusterOptionsFlux() OptionsFlux {
	return OptionsFlux{}
}

// NewClusterOptionsArgoCD creates a new OptionsArgoCD with default values.
func NewClusterOptionsArgoCD() OptionsArgoCD {
	return OptionsArgoCD{}
}

// NewClusterOptionsHelm creates a new OptionsHelm with default values.
func NewClusterOptionsHelm() OptionsHelm {
	return OptionsHelm{}
}

// NewClusterOptionsKustomize creates a new OptionsKustomize with default values.
func NewClusterOptionsKustomize() OptionsKustomize {
	return OptionsKustomize{}
}
