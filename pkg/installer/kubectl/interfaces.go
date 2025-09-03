package kubectlinstaller

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

// APIExtensionsClientInterface defines the interface for API extensions client operations.
type APIExtensionsClientInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Create(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition, opts metav1.CreateOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Update(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition, opts metav1.UpdateOptions) (*apiextensionsv1.CustomResourceDefinition, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

// DynamicClientInterface defines the interface for dynamic client operations.
type DynamicClientInterface interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions) (*unstructured.Unstructured, error)
	Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

// ClientFactoryInterface defines the interface for creating Kubernetes clients.
type ClientFactoryInterface interface {
	CreateAPIExtensionsClient(config *rest.Config) (APIExtensionsClientInterface, error)
	CreateDynamicClient(config *rest.Config, gvr schema.GroupVersionResource) (DynamicClientInterface, error)
}