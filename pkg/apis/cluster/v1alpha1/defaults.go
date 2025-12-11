package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultDistributionConfig is the default config file name for Kind clusters.
	DefaultDistributionConfig = "kind.yaml"
	// DefaultK3dDistributionConfig is the default config file name for K3d clusters.
	DefaultK3dDistributionConfig = "k3d.yaml"
	// DefaultSourceDirectory is the default directory for Kubernetes manifests.
	DefaultSourceDirectory = "k8s"
	// DefaultKubeconfigPath is the default path to the kubeconfig file.
	DefaultKubeconfigPath = "~/.kube/config"
	// DefaultLocalRegistryPort is the default port for the local OCI registry.
	DefaultLocalRegistryPort int32 = 5111
)

// DefaultFluxInterval is the default reconciliation interval for Flux resources.
//
//nolint:gochecknoglobals // This is a legitimate package-level default constant value.
var DefaultFluxInterval = metav1.Duration{Duration: time.Minute}

// ExpectedDistributionConfigName returns the default config filename for a distribution.
func ExpectedDistributionConfigName(distribution Distribution) string {
	switch distribution {
	case DistributionKind:
		return DefaultDistributionConfig
	case DistributionK3d:
		return DefaultK3dDistributionConfig
	default:
		return DefaultDistributionConfig
	}
}

// ExpectedContextName returns the default kube context name for a distribution.
func ExpectedContextName(distribution Distribution) string {
	switch distribution {
	case DistributionKind:
		return "kind-kind"
	case DistributionK3d:
		return "k3d-k3d-default"
	default:
		return ""
	}
}
