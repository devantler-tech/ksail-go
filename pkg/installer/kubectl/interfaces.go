package kubectlinstaller

import (
	"context"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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

// defaultAPIExtensionsClient wraps the real API extensions client.
type defaultAPIExtensionsClient struct {
	client *apiextensionsclient.Clientset
}

func (c *defaultAPIExtensionsClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
	return c.client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, opts)
}

func (c *defaultAPIExtensionsClient) Create(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition, opts metav1.CreateOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
	return c.client.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, opts)
}

func (c *defaultAPIExtensionsClient) Update(ctx context.Context, crd *apiextensionsv1.CustomResourceDefinition, opts metav1.UpdateOptions) (*apiextensionsv1.CustomResourceDefinition, error) {
	return c.client.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, crd, opts)
}

func (c *defaultAPIExtensionsClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, name, opts)
}

// defaultDynamicClient wraps the real dynamic client.
type defaultDynamicClient struct {
	client dynamic.Interface
	gvr    schema.GroupVersionResource
}

func (c *defaultDynamicClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	return c.client.Resource(c.gvr).Get(ctx, name, opts)
}

func (c *defaultDynamicClient) Create(ctx context.Context, obj *unstructured.Unstructured, opts metav1.CreateOptions) (*unstructured.Unstructured, error) {
	return c.client.Resource(c.gvr).Create(ctx, obj, opts)
}

func (c *defaultDynamicClient) Update(ctx context.Context, obj *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	return c.client.Resource(c.gvr).Update(ctx, obj, opts)
}

func (c *defaultDynamicClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Resource(c.gvr).Delete(ctx, name, opts)
}

// defaultClientFactory creates real Kubernetes clients.
type defaultClientFactory struct{}

func (f *defaultClientFactory) CreateAPIExtensionsClient(config *rest.Config) (APIExtensionsClientInterface, error) {
	client, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &defaultAPIExtensionsClient{client: client}, nil
}

func (f *defaultClientFactory) CreateDynamicClient(config *rest.Config, gvr schema.GroupVersionResource) (DynamicClientInterface, error) {
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &defaultDynamicClient{client: client, gvr: gvr}, nil
}

// NewDefaultClientFactory creates a new default client factory.
func NewDefaultClientFactory() ClientFactoryInterface {
	return &defaultClientFactory{}
}