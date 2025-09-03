// Package kubectlinstaller provides interfaces for kubectl installer dependencies.
package kubectlinstaller

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

// ClientFactory defines the interface for creating Kubernetes clients.
type ClientFactory interface {
	CreateAPIExtensionsClient(config *rest.Config) (APIExtensionsClient, error)
	CreateDynamicClient(config *rest.Config, gvr schema.GroupVersionResource) (DynamicClient, error)
}