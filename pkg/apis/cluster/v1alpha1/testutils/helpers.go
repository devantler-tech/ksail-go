// Package testutils provides common test utilities for cluster API v1alpha1 types.
package testutils

import (
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateDefaultObjectMeta creates a default metav1.ObjectMeta for testing.
func CreateDefaultObjectMeta(name string) metav1.ObjectMeta {
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
		Labels:                     map[string]string{},
		Annotations:                map[string]string{},
		OwnerReferences:            []metav1.OwnerReference{},
		Finalizers:                 []string{},
		ManagedFields:              []metav1.ManagedFieldsEntry{},
	}
}

// CreateDefaultSpecOptions creates a default v1alpha1.Options for testing.
func CreateDefaultSpecOptions() v1alpha1.Options {
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
		Options:            CreateDefaultSpecOptions(),
	}
}
