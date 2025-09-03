// Package defaults provides default configuration values for K3d cluster generation.
package defaults

import v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"

// K3dOptions returns default K3d options for cluster configuration.
func K3dOptions() v1alpha5.SimpleConfigOptionsK3d {
	return v1alpha5.SimpleConfigOptionsK3d{
		Wait:                false,
		Timeout:             0,
		DisableLoadbalancer: false,
		DisableImageVolume:  false,
		NoRollback:          false,
		NodeHookActions:     nil,
		Loadbalancer: v1alpha5.SimpleConfigOptionsK3dLoadbalancer{
			ConfigOverrides: nil,
		},
	}
}