package kubectl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Interface defines operations for applying Kubernetes manifests from URLs.
// This interface is used for programmatic manifest application, particularly by CNI installers.
type Interface interface {
	// Apply fetches a manifest from a URL and applies it to the cluster.
	Apply(ctx context.Context, manifestURL string) error
	// Delete removes a specific resource from the cluster.
	Delete(ctx context.Context, namespace string, resourceType string, name string) error
}

// ManifestClient implements kubectl-like manifest operations using Kubernetes dynamic client.
type ManifestClient struct {
	kubeconfig      string
	context         string
	restConfig      *rest.Config
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	mapper          meta.RESTMapper
}

const applyFieldManager = "ksail-kubectl-client"

var (
	errEmptyResourceIdentifiers = errors.New("resource type and name must not be empty")
	errInvalidResourceType      = errors.New("invalid resource type")
	errUnexpectedStatusCode     = errors.New("unexpected status code")
)

// NewManifestClient creates a new kubectl manifest client instance.
//
//nolint:ireturn // Interface return type is required for dependency injection.
func NewManifestClient(kubeconfig, context string) (Interface, error) {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Create discovery client for RESTMapper
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery client: %w", err)
	}

	// Create cached discovery and RESTMapper
	cachedDiscovery := memory.NewMemCacheClient(discoveryClient)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscovery)

	return &ManifestClient{
		kubeconfig:      kubeconfig,
		context:         context,
		restConfig:      config,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		mapper:          mapper,
	}, nil
}

// Apply fetches a manifest from URL and applies all resources to the cluster.
func (c *ManifestClient) Apply(ctx context.Context, manifestURL string) error {
	// Fetch manifest from URL
	manifestYAML, err := c.fetchManifest(ctx, manifestURL)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest from %s: %w", manifestURL, err)
	}

	// Split YAML document by ---  separator
	docs := strings.Split(string(manifestYAML), "\n---\n")

	// Apply each document
	decoder := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	for docIndex, rawDoc := range docs {
		document := strings.TrimSpace(rawDoc)
		if document == "" {
			continue
		}

		obj := &unstructured.Unstructured{}

		_, gvk, err := decoder.Decode([]byte(document), nil, obj)
		if err != nil {
			return fmt.Errorf("failed to decode manifest document %d: %w", docIndex, err)
		}

		if obj.GetKind() == "" {
			continue
		}

		err = c.applyResource(ctx, obj, gvk)
		if err != nil {
			return fmt.Errorf(
				"failed to apply resource %s/%s (document %d): %w",
				obj.GetKind(),
				obj.GetName(),
				docIndex,
				err,
			)
		}
	}

	return nil
}

// Delete removes a specific resource from the cluster.
func (c *ManifestClient) Delete(
	ctx context.Context,
	namespace string,
	resourceType string,
	name string,
) error {
	err := validateResourceIdentifiers(resourceType, name)
	if err != nil {
		return err
	}

	mapping, err := c.resolveResourceMapping(resourceType)
	if err != nil {
		return err
	}

	var resourceClient dynamic.ResourceInterface

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resolvedNamespace := resolveNamespace(namespace)
		resourceClient = c.dynamicClient.Resource(mapping.Resource).Namespace(resolvedNamespace)
	} else {
		resourceClient = c.dynamicClient.Resource(mapping.Resource)
	}

	err = resourceClient.Delete(ctx, name, metav1.DeleteOptions{})
	if apierrors.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to delete %s %q: %w", resourceType, name, err)
	}

	return nil
}

// applyResource applies a single unstructured resource to the cluster.
func (c *ManifestClient) applyResource(
	ctx context.Context,
	obj *unstructured.Unstructured,
	gvk *schema.GroupVersionKind,
) error {
	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return fmt.Errorf("failed to get REST mapping for %s: %w", gvk, err)
	}

	var resourceClient dynamic.ResourceInterface

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		resolvedNamespace := resolveNamespace(obj.GetNamespace())
		resourceClient = c.dynamicClient.Resource(mapping.Resource).Namespace(resolvedNamespace)
	} else {
		resourceClient = c.dynamicClient.Resource(mapping.Resource)
	}

	data, err := obj.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal object: %w", err)
	}

	_, err = resourceClient.Patch(
		ctx,
		obj.GetName(),
		types.ApplyPatchType,
		data,
		metav1.PatchOptions{FieldManager: applyFieldManager},
	)
	if apierrors.IsNotFound(err) {
		_, createErr := resourceClient.Create(ctx, obj, metav1.CreateOptions{})
		if createErr != nil {
			return fmt.Errorf("failed to create resource during apply fallback: %w", createErr)
		}

		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to apply resource: %w", err)
	}

	return nil
}

func validateResourceIdentifiers(resourceType, name string) error {
	if resourceType == "" || name == "" {
		return fmt.Errorf(
			"%w: resourceType=%q name=%q",
			errEmptyResourceIdentifiers,
			resourceType,
			name,
		)
	}

	return nil
}

func resolveNamespace(namespace string) string {
	if namespace != "" {
		return namespace
	}

	return "default"
}

func (c *ManifestClient) resolveResourceMapping(resourceType string) (*meta.RESTMapping, error) {
	gvr, err := c.resolveGroupVersionResource(resourceType)
	if err != nil {
		return nil, err
	}

	gvk, err := c.mapper.KindFor(gvr)
	if err != nil {
		return nil, fmt.Errorf("failed to determine kind for %s: %w", gvr.String(), err)
	}

	mapping, err := c.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping for %s: %w", gvk.String(), err)
	}

	return mapping, nil
}

func (c *ManifestClient) resolveGroupVersionResource(
	resourceType string,
) (schema.GroupVersionResource, error) {
	parsedGVR, groupResource := schema.ParseResourceArg(resourceType)

	switch {
	case parsedGVR != nil:
		gvr, err := c.mapper.ResourceFor(*parsedGVR)
		if err != nil {
			return schema.GroupVersionResource{}, fmt.Errorf(
				"failed to resolve resource type %q: %w",
				resourceType,
				err,
			)
		}

		return gvr, nil
	case !groupResource.Empty():
		partial := schema.GroupVersionResource{
			Group:    groupResource.Group,
			Resource: groupResource.Resource,
		}

		gvr, err := c.mapper.ResourceFor(partial)
		if err != nil {
			return schema.GroupVersionResource{}, fmt.Errorf(
				"failed to resolve resource type %q: %w",
				resourceType,
				err,
			)
		}

		return gvr, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf(
			"%w: %s",
			errInvalidResourceType,
			resourceType,
		)
	}
}

// fetchManifest downloads manifest content from URL.
func (c *ManifestClient) fetchManifest(
	ctx context.Context,
	manifestURL string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch manifest: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"%w %d from manifest URL %s",
			errUnexpectedStatusCode,
			resp.StatusCode,
			manifestURL,
		)
	}

	manifestBytes, err := io.ReadAll(resp.Body)

	closeErr := resp.Body.Close()
	if err != nil {
		if closeErr != nil {
			err = errors.Join(err, fmt.Errorf("close response body: %w", closeErr))
		}

		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("close response body: %w", closeErr)
	}

	return manifestBytes, nil
}
