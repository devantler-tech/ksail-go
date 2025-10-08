// Package k8sclient provides utilities for creating Kubernetes clients.
package k8sclient

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ClientProvider describes a provider that can create Kubernetes clients.
type ClientProvider interface {
	// CreateClient creates a Kubernetes clientset from the given kubeconfig and context.
	CreateClient(kubeconfig, context string) (*kubernetes.Clientset, error)
}

// DefaultClientProvider is the default implementation of ClientProvider.
type DefaultClientProvider struct{}

// NewDefaultClientProvider creates a new DefaultClientProvider.
func NewDefaultClientProvider() *DefaultClientProvider {
	return &DefaultClientProvider{}
}

// CreateClient creates a Kubernetes clientset from the given kubeconfig and context.
// If kubeconfig is empty, it uses the default location (~/.kube/config) or in-cluster config.
// If context is empty, it uses the current context from the kubeconfig.
func (p *DefaultClientProvider) CreateClient(
	kubeconfig, context string,
) (*kubernetes.Clientset, error) {
	config, err := p.buildConfig(kubeconfig, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, nil
}

// buildConfig builds a Kubernetes rest.Config from kubeconfig file and context.
func (p *DefaultClientProvider) buildConfig(kubeconfig, context string) (*rest.Config, error) {
	// If kubeconfig is not provided, try default location
	if kubeconfig == "" {
		kubeconfig = p.getDefaultKubeconfig()
	}

	// Try to use kubeconfig file if it exists
	_, statErr := os.Stat(kubeconfig)
	if statErr == nil {
		return p.buildConfigFromKubeconfig(kubeconfig, context)
	}

	// Fall back to in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	return config, nil
}

// buildConfigFromKubeconfig builds config from a kubeconfig file.
func (p *DefaultClientProvider) buildConfigFromKubeconfig(
	kubeconfig, context string,
) (*rest.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{}

	if context != "" {
		configOverrides.CurrentContext = context
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	return config, nil
}

// getDefaultKubeconfig returns the default kubeconfig location.
func (p *DefaultClientProvider) getDefaultKubeconfig() string {
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}
