package configmanager

import (
	"fmt"

	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func GetClusterName(any any) (string, error) {
	switch cfg := any.(type) {
	case *v1alpha4.Cluster:
		return cfg.Name, nil
	case *v1alpha5.SimpleConfig:
		return cfg.Name, nil
	default:
		return "", fmt.Errorf("unsupported config type: '%T'", cfg)
	}
}
