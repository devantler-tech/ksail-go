// Package testutils provides common test utilities for K3D provisioner tests.
package testutils

import (
	"net/netip"

	"github.com/docker/go-connections/nat"
	"github.com/k3d-io/k3d/v5/pkg/types"
	configtypes "github.com/k3d-io/k3d/v5/pkg/config/types"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	wharfie "github.com/rancher/wharfie/pkg/registries"
)

// CreateDefaultK3dOptions creates a default SimpleConfigOptionsK3d for testing.
func CreateDefaultK3dOptions() v1alpha5.SimpleConfigOptionsK3d {
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

// CreateDefaultCluster creates a default types.Cluster for testing.
func CreateDefaultCluster(name string) *types.Cluster {
	return &types.Cluster{
		Name: name,
		Network: types.ClusterNetwork{
			Name:     "",
			ID:       "",
			External: false,
			IPAM: types.IPAM{
				IPPrefix: netip.Prefix{},
				IPsUsed:  nil,
				Managed:  false,
			},
			Members: nil,
		},
		Token:              "",
		Nodes:              nil,
		InitNode:           nil,
		ExternalDatastore:  nil,
		KubeAPI:            nil,
		ServerLoadBalancer: nil,
		ImageVolume:        "",
		Volumes:            nil,
	}
}

// CreateClusterWithKubeAPI creates a cluster with KubeAPI configuration for testing.
func CreateClusterWithKubeAPI(name string) *types.Cluster {
	cluster := CreateDefaultCluster(name)
	cluster.KubeAPI = &types.ExposureOpts{
		PortMapping: nat.PortMapping{
			Port: "",
			Binding: nat.PortBinding{
				HostIP:   "",
				HostPort: "",
			},
		},
		Host: "",
	}
	cluster.ServerLoadBalancer = &types.Loadbalancer{
		Node:   nil,
		Config: nil,
	}
	return cluster
}

// CreateDefaultClusterConfig creates a default v1alpha5.ClusterConfig for testing.
func CreateDefaultClusterConfig() *v1alpha5.ClusterConfig {
	return &v1alpha5.ClusterConfig{
		TypeMeta: configtypes.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		Cluster: *CreateDefaultCluster(""),
		ClusterCreateOpts: types.ClusterCreateOpts{
			DisableImageVolume:   false,
			WaitForServer:        false,
			Timeout:              0,
			DisableLoadBalancer:  false,
			GPURequest:           "",
			ServersMemory:        "",
			AgentsMemory:         "",
			NodeHooks:            nil,
			GlobalLabels:         nil,
			GlobalEnv:            nil,
			HostAliases:          nil,
			Registries: struct {
				Create *types.Registry                        `json:"create,omitempty"`
				Use    []*types.Registry                      `json:"use,omitempty"`
				Config *wharfie.Registry `json:"config,omitempty"`
			}{
				Create: nil,
				Use:    nil,
				Config: nil,
			},
		},
		KubeconfigOpts: v1alpha5.SimpleConfigOptionsKubeconfig{
			UpdateDefaultKubeconfig: false,
			SwitchCurrentContext:    false,
		},
	}
}