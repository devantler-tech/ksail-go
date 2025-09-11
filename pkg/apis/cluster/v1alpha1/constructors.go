// Package v1alpha1 provides model definitions for a KSail cluster.
package v1alpha1

import (
	"time"

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
		Metadata: metav1.ObjectMeta{}, //nolint:exhaustruct // Intentionally empty, filled in later
		Spec:     Spec{},              //nolint:exhaustruct // Intentionally empty, filled in later
	}
}

// NewClusterMetadata creates a new metav1.ObjectMeta with the specified name and default values.
func NewClusterMetadata(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:                       name,
		GenerateName:               "",
		Namespace:                  "",
		SelfLink:                   "",
		UID:                        "",
		ResourceVersion:            "",
		Generation:                 0,
		CreationTimestamp:          metav1.Time{Time: time.Time{}},
		DeletionTimestamp:          nil,
		DeletionGracePeriodSeconds: nil,
		Labels:                     nil,
		Annotations:                nil,
		OwnerReferences:            []metav1.OwnerReference{},
		Finalizers:                 []string{},
		ManagedFields:              []metav1.ManagedFieldsEntry{},
	}
}

// NewSpec creates a new Spec with default values.
func NewSpec() Spec {
	return Spec{
		DistributionConfig: "",
		SourceDirectory:    "",
		Connection:         NewConnection(),
		Distribution:       "",
		CNI:                "",
		CSI:                "",
		IngressController:  "",
		GatewayController:  "",
		ReconciliationTool: "",
		Options:            NewOptions(),
	}
}

// NewConnection creates a new Connection with default values.
func NewConnection() Connection {
	return Connection{
		Kubeconfig: "",
		Context:    "",
		Timeout:    metav1.Duration{Duration: 0},
	}
}

// NewOptions creates a new Options with default values.
func NewOptions() Options {
	return Options{
		Kind:      NewOptionsKind(),
		K3d:       NewOptionsK3d(),
		Tind:      NewOptionsTind(),
		EKS:       NewOptionsEKS(),
		Cilium:    NewOptionsCilium(),
		Kubectl:   NewOptionsKubectl(),
		Flux:      NewOptionsFlux(),
		ArgoCD:    NewOptionsArgoCD(),
		Helm:      NewOptionsHelm(),
		Kustomize: NewOptionsKustomize(),
	}
}

// NewOptionsKind creates a new OptionsKind with default values.
func NewOptionsKind() OptionsKind {
	return OptionsKind{}
}

// NewOptionsK3d creates a new OptionsK3d with default values.
func NewOptionsK3d() OptionsK3d {
	return OptionsK3d{}
}

// NewOptionsTind creates a new OptionsTind with default values.
func NewOptionsTind() OptionsTind {
	return OptionsTind{}
}

// NewOptionsEKS creates a new OptionsEKS with default values.
func NewOptionsEKS() OptionsEKS {
	return OptionsEKS{
		AWSProfile: "",
	}
}

// NewOptionsCilium creates a new OptionsCilium with default values.
func NewOptionsCilium() OptionsCilium {
	return OptionsCilium{}
}

// NewOptionsKubectl creates a new OptionsKubectl with default values.
func NewOptionsKubectl() OptionsKubectl {
	return OptionsKubectl{}
}

// NewOptionsFlux creates a new OptionsFlux with default values.
func NewOptionsFlux() OptionsFlux {
	return OptionsFlux{}
}

// NewOptionsArgoCD creates a new OptionsArgoCD with default values.
func NewOptionsArgoCD() OptionsArgoCD {
	return OptionsArgoCD{}
}

// NewOptionsHelm creates a new OptionsHelm with default values.
func NewOptionsHelm() OptionsHelm {
	return OptionsHelm{}
}

// NewOptionsKustomize creates a new OptionsKustomize with default values.
func NewOptionsKustomize() OptionsKustomize {
	return OptionsKustomize{}
}
