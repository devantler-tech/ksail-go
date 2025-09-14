// Package ksail provides configuration management for KSail v1alpha1.Cluster configurations.
// This file contains the core Manager implementation.
package ksail

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Manager implements the ConfigManager interface for KSail v1alpha1.Cluster configurations.
type Manager struct {
	viper          *viper.Viper
	fieldSelectors []FieldSelector[v1alpha1.Cluster]
	Config         *v1alpha1.Cluster // Exposed config property as suggested
}

// Verify that Manager implements the ConfigManager interface.
var _ configmanager.ConfigManager[v1alpha1.Cluster] = (*Manager)(nil)

// NewManager creates a new configuration manager with the specified field selectors.
func NewManager(fieldSelectors ...FieldSelector[v1alpha1.Cluster]) *Manager {
	return &Manager{
		viper:          InitializeViper(),
		fieldSelectors: fieldSelectors,
		Config: &v1alpha1.Cluster{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.Kind,
				APIVersion: v1alpha1.APIVersion,
			},
			Metadata: metav1.ObjectMeta{
				Name:            "",
				GenerateName:    "",
				Namespace:       "",
				SelfLink:        "",
				UID:             "",
				ResourceVersion: "",
				Generation:      0,
				CreationTimestamp: metav1.Time{
					Time: time.Time{},
				},
				DeletionTimestamp:          nil,
				DeletionGracePeriodSeconds: nil,
				Labels:                     nil,
				Annotations:                nil,
				OwnerReferences:            nil,
				Finalizers:                 nil,
				ManagedFields:              nil,
			},
			Spec: v1alpha1.Spec{
				DistributionConfig: "",
				SourceDirectory:    "",
				Connection: v1alpha1.Connection{
					Kubeconfig: "",
					Context:    "",
					Timeout: metav1.Duration{
						Duration: 0,
					},
				},
				Distribution:       "",
				CNI:                "",
				CSI:                "",
				IngressController:  "",
				GatewayController:  "",
				ReconciliationTool: "",
				Options: v1alpha1.Options{
					Kind: v1alpha1.OptionsKind{},
					K3d:  v1alpha1.OptionsK3d{},
					Tind: v1alpha1.OptionsTind{},
					EKS: v1alpha1.OptionsEKS{
						AWSProfile: "",
					},
					Cilium:    v1alpha1.OptionsCilium{},
					Kubectl:   v1alpha1.OptionsKubectl{},
					Flux:      v1alpha1.OptionsFlux{},
					ArgoCD:    v1alpha1.OptionsArgoCD{},
					Helm:      v1alpha1.OptionsHelm{},
					Kustomize: v1alpha1.OptionsKustomize{},
				},
			},
		},
	}
}

// LoadConfig loads the configuration from files and environment variables.
// Returns the previously loaded config if already loaded.
func (m *Manager) LoadConfig() (*v1alpha1.Cluster, error) {
	// If config is already loaded and populated, return it
	if m.Config != nil && !isEmptyCluster(m.Config) {
		return m.Config, nil
	}

	// Initialize with defaults from field selectors
	m.applyDefaults()

	// Try to read from configuration files
	m.viper.SetConfigName(DefaultConfigFileName)
	m.viper.SetConfigType("yaml")
	m.viper.AddConfigPath(".")
	m.viper.AddConfigPath("$HOME/.config/ksail")
	m.viper.AddConfigPath("/etc/ksail")

	// Read configuration file if it exists
	err := m.viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults and flags
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Set environment variable prefix and bind environment variables
	m.viper.SetEnvPrefix(EnvPrefix)
	m.viper.AutomaticEnv()
	bindEnvironmentVariables(m.viper)

	// Unmarshal into our cluster config
	err = m.viper.Unmarshal(m.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return m.Config, nil
}

// isEmptyCluster checks if the cluster configuration is empty/default.
func isEmptyCluster(config *v1alpha1.Cluster) bool {
	emptyCluster := &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.Kind,
			APIVersion: v1alpha1.APIVersion,
		},
		Metadata: metav1.ObjectMeta{
			Name:            "",
			GenerateName:    "",
			Namespace:       "",
			SelfLink:        "",
			UID:             "",
			ResourceVersion: "",
			Generation:      0,
			CreationTimestamp: metav1.Time{
				Time: time.Time{},
			},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     nil,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Spec: v1alpha1.Spec{
			DistributionConfig: "",
			SourceDirectory:    "",
			Connection: v1alpha1.Connection{
				Kubeconfig: "",
				Context:    "",
				Timeout: metav1.Duration{
					Duration: 0,
				},
			},
			Distribution:       "",
			CNI:                "",
			CSI:                "",
			IngressController:  "",
			GatewayController:  "",
			ReconciliationTool: "",
			Options: v1alpha1.Options{
				Kind: v1alpha1.OptionsKind{},
				K3d:  v1alpha1.OptionsK3d{},
				Tind: v1alpha1.OptionsTind{},
				EKS: v1alpha1.OptionsEKS{
					AWSProfile: "",
				},
				Cilium:    v1alpha1.OptionsCilium{},
				Kubectl:   v1alpha1.OptionsKubectl{},
				Flux:      v1alpha1.OptionsFlux{},
				ArgoCD:    v1alpha1.OptionsArgoCD{},
				Helm:      v1alpha1.OptionsHelm{},
				Kustomize: v1alpha1.OptionsKustomize{},
			},
		},
	}

	return reflect.DeepEqual(config, emptyCluster)
}

// GetViper returns the underlying Viper instance for flag binding.
func (m *Manager) GetViper() *viper.Viper {
	return m.viper
}

// applyDefaults applies default values from field selectors to the config.
func (m *Manager) applyDefaults() {
	for _, fieldSelector := range m.fieldSelectors {
		fieldPtr := fieldSelector.Selector(m.Config)
		if fieldPtr != nil {
			setFieldValue(fieldPtr, fieldSelector.DefaultValue)
		}
	}
}
