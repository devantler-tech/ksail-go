package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
// DefaultDistributionConfig is the default cluster distribution configuration filename.
	DefaultDistributionConfig          = "kind.yaml"
	DefaultK3dDistributionConfig       = "k3d.yaml"
	DefaultSourceDirectory             = "k8s"
	DefaultKubeconfigPath              = "~/.kube/config"
	DefaultLocalRegistryPort     int32 = 5111
)

var (
	// DefaultFluxInterval is the default reconciliation interval for Flux.
	DefaultFluxInterval = metav1.Duration{Duration: time.Minute}
)

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
