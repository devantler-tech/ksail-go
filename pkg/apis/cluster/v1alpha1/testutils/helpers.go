// Package testutils provides common test utilities for cluster API v1alpha1 types.
package testutils

import (
	"time"

	"github.com/devantler-tech/ksail-go/internal/k8s"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateDefaultObjectMeta creates a default metav1.ObjectMeta for testing.
func CreateDefaultObjectMeta(name string) metav1.ObjectMeta {
	return k8s.NewObjectMeta(name)
}

// CreateDefaultOptions creates a default v1alpha1.Options for testing.
func CreateDefaultOptions() v1alpha1.Options {
	return v1alpha1.Options{
		Kind:      v1alpha1.OptionsKind{},
		K3d:       v1alpha1.OptionsK3d{},
		Tind:      v1alpha1.OptionsTind{},
		Cilium:    v1alpha1.OptionsCilium{},
		Kubectl:   v1alpha1.OptionsKubectl{},
		Flux:      v1alpha1.OptionsFlux{},
		ArgoCD:    v1alpha1.OptionsArgoCD{},
		Helm:      v1alpha1.OptionsHelm{},
		Kustomize: v1alpha1.OptionsKustomize{},
	}
}

// CreateDefaultSpec creates a default v1alpha1.Spec for testing.
func CreateDefaultSpec() v1alpha1.Spec {
	return v1alpha1.Spec{
		Distribution:       "",
		DistributionConfig: "",
		SourceDirectory:    "",
		Connection: v1alpha1.Connection{
			Kubeconfig: "",
			Context:    "",
			Timeout:    metav1.Duration{Duration: time.Duration(0)},
		},
		ContainerEngine:    "",
		CNI:                "",
		CSI:                "",
		IngressController:  "",
		GatewayController:  "",
		ReconciliationTool: "",
		Options:            CreateDefaultOptions(),
	}
}

// CreateDefaultK3dSpec creates a default v1alpha1.Spec configured for K3d for testing.
func CreateDefaultK3dSpec() v1alpha1.Spec {
	spec := CreateDefaultSpec()
	spec.Distribution = v1alpha1.DistributionK3d
	return spec
}