package kubectl //nolint:testpackage // Tests require access to internal client structure for fake setup.

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	apiMeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
)

// newTestRESTMapper builds a RESTMapper with the resources required for tests.
//
//nolint:ireturn // Tests need RESTMapper interface for fake dynamic client.
func newTestRESTMapper() apiMeta.RESTMapper {
	versions := []schema.GroupVersion{
		{Group: "", Version: "v1"},
		{Group: "apps", Version: "v1"},
	}

	mapper := apiMeta.NewDefaultRESTMapper(versions)
	mapper.AddSpecific(
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
		schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmap"},
		apiMeta.RESTScopeNamespace,
	)
	mapper.AddSpecific(
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"},
		schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonset"},
		apiMeta.RESTScopeNamespace,
	)
	mapper.AddSpecific(
		schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
		schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"},
		schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespace"},
		apiMeta.RESTScopeRoot,
	)

	return mapper
}

// newManifestClientForTest creates a ManifestClient configured with fake dependencies.
func newManifestClientForTest(
	t *testing.T,
	objects ...runtime.Object,
) (*ManifestClient, *fake.FakeDynamicClient) {
	t.Helper()

	scheme := runtime.NewScheme()
	require.NoError(t, v1.AddToScheme(scheme))

	dynamicClient := fake.NewSimpleDynamicClient(scheme, objects...)

	client := &ManifestClient{
		dynamicClient: dynamicClient,
		mapper:        newTestRESTMapper(),
	}

	return client, dynamicClient
}

func TestManifestClient_Apply_Succeeds(t *testing.T) {
	t.Parallel()

	manifest := strings.TrimSpace(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
  namespace: test-ns
data:
  key: value
`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, manifest)
	}))
	t.Cleanup(server.Close)

	client, dynamicClient := newManifestClientForTest(t)

	err := client.Apply(context.Background(), server.URL)
	require.NoError(t, err)

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	applied, err := dynamicClient.Resource(gvr).
		Namespace("test-ns").
		Get(context.Background(), "test-config", metav1.GetOptions{})
	require.NoError(t, err)

	data, ok := applied.Object["data"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "value", data["key"])
}

func TestManifestClient_Apply_PropagatesFetchError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client, _ := newManifestClientForTest(t)

	err := client.Apply(context.Background(), server.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected status code 500")
}

func TestManifestClient_Delete_RemovesResource(t *testing.T) {
	t.Parallel()

	existing := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "test-ns",
		},
		Data: map[string]string{"key": "value"},
	}

	client, dynamicClient := newManifestClientForTest(t, existing)

	err := client.Delete(context.Background(), "test-ns", "configmap", "test-config")
	require.NoError(t, err)

	gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	_, err = dynamicClient.Resource(gvr).
		Namespace("test-ns").
		Get(context.Background(), "test-config", metav1.GetOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManifestClient_Delete_IdempotentWhenResourceMissing(t *testing.T) {
	t.Parallel()

	client, _ := newManifestClientForTest(t)

	err := client.Delete(context.Background(), "test-ns", "configmap", "missing")
	require.NoError(t, err)
}
