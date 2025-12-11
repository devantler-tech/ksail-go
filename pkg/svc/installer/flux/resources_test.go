//nolint:testpackage // test needs access to unexported functions
package fluxinstaller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	fluxclient "github.com/devantler-tech/ksail-go/pkg/client/flux"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	discoveryfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/rest"
	clientgotesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestBuildFluxInstanceUsesDefaults(t *testing.T) {
	t.Parallel()

	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			SourceDirectory: "workloads/env/prod",
			Options:         v1alpha1.Options{},
			LocalRegistry:   v1alpha1.LocalRegistryEnabled,
		},
	}

	instance, err := buildFluxInstance(cluster)
	require.NoError(t, err)
	require.Equal(t, fluxInstanceDefaultName, instance.Name)
	require.Equal(t, fluxclient.DefaultNamespace, instance.Namespace)
	require.Equal(t, "oci://local-registry:5000/workloads-env-prod", instance.Spec.Sync.URL)
	require.Equal(t, "./", instance.Spec.Sync.Path)
	require.Equal(t, defaultArtifactTag, instance.Spec.Sync.Ref)
	require.Equal(t, metav1.Duration{Duration: fluxIntervalFallback}, *instance.Spec.Sync.Interval)
}

func TestBuildFluxInstanceRespectsClusterOptions(t *testing.T) {
	t.Parallel()

	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			SourceDirectory: " ../My Workloads  ",
			Options: v1alpha1.Options{
				Flux:          v1alpha1.OptionsFlux{Interval: metav1.Duration{Duration: 2 * time.Minute}},
				LocalRegistry: v1alpha1.OptionsLocalRegistry{HostPort: 5111},
			},
			LocalRegistry: v1alpha1.LocalRegistryEnabled,
		},
	}

	instance, err := buildFluxInstance(cluster)
	require.NoError(t, err)
	require.Equal(t, "oci://local-registry:5000/my-workloads", instance.Spec.Sync.URL)
	require.Equal(t, metav1.Duration{Duration: 2 * time.Minute}, *instance.Spec.Sync.Interval)
	require.Equal(t, "./", instance.Spec.Sync.Path)
}

func TestBuildFluxInstanceFallsBackWhenRegistryDisabled(t *testing.T) {
	t.Parallel()

	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			SourceDirectory: "k8s",
			Options: v1alpha1.Options{
				LocalRegistry: v1alpha1.OptionsLocalRegistry{HostPort: 5999},
			},
			LocalRegistry: v1alpha1.LocalRegistryDisabled,
		},
	}

	instance, err := buildFluxInstance(cluster)
	require.NoError(t, err)
	require.Equal(t, "oci://localhost:5999/k8s", instance.Spec.Sync.URL)
}

func TestEnsureDefaultResourcesCreatesAndUpdatesFluxInstance(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	require.NoError(t, addFluxInstanceToScheme(scheme))
	require.NoError(t, sourcev1.AddToScheme(scheme))

	defaultOCIRepo := &sourcev1.OCIRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      defaultOCIRepositoryName,
			Namespace: fluxclient.DefaultNamespace,
		},
	}

	fakeClient := crfake.NewClientBuilder().WithScheme(scheme).WithObjects(defaultOCIRepo).Build()

	overrideRESTConfigLoader(t, func(string) (*rest.Config, error) {
		return &rest.Config{}, nil
	})
	stubDiscovery := newStubDiscovery(nil)

	overrideDiscoveryClientFactory(t, func(*rest.Config) (discovery.DiscoveryInterface, error) {
		return stubDiscovery, nil
	})
	overrideFluxResourcesClientFactory(t, func(*rest.Config) (client.Client, error) {
		return fakeClient, nil
	})

	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.Spec{
			SourceDirectory: "k8s",
			Options: v1alpha1.Options{
				Flux:          v1alpha1.OptionsFlux{Interval: metav1.Duration{Duration: time.Minute}},
				LocalRegistry: v1alpha1.OptionsLocalRegistry{HostPort: 5001},
			},
			LocalRegistry: v1alpha1.LocalRegistryEnabled,
		},
	}

	require.NoError(t, EnsureDefaultResources(context.Background(), "kubeconfig", cluster))

	instance := &FluxInstance{}
	key := client.ObjectKey{Name: fluxInstanceDefaultName, Namespace: fluxclient.DefaultNamespace}
	require.NoError(t, fakeClient.Get(context.Background(), key, instance))
	require.Equal(t, metav1.Duration{Duration: time.Minute}, *instance.Spec.Sync.Interval)
	require.Equal(t, "oci://local-registry:5000/k8s", instance.Spec.Sync.URL)

	repo := &sourcev1.OCIRepository{}
	repoKey := client.ObjectKey{Name: defaultOCIRepositoryName, Namespace: fluxclient.DefaultNamespace}
	require.NoError(t, fakeClient.Get(context.Background(), repoKey, repo))
	require.True(t, repo.Spec.Insecure)

	cluster.Spec.Options.Flux.Interval = metav1.Duration{Duration: 3 * time.Minute}

	require.NoError(t, EnsureDefaultResources(context.Background(), "kubeconfig", cluster))
	require.NoError(t, fakeClient.Get(context.Background(), key, instance))
	require.Equal(t, metav1.Duration{Duration: 3 * time.Minute}, *instance.Spec.Sync.Interval)
	require.GreaterOrEqual(t, stubDiscovery.calls, 2)
}

func TestEnsureDefaultResourcesFailsWhenFluxInstanceAPIsUnavailable(t *testing.T) {
	t.Parallel()

	setFluxAPITimeouts(t, 15*time.Millisecond, time.Millisecond)

	overrideRESTConfigLoader(t, func(string) (*rest.Config, error) {
		return &rest.Config{}, nil
	})

	overrideDiscoveryClientFactory(t, func(*rest.Config) (discovery.DiscoveryInterface, error) {
		return newStubDiscovery(map[string]error{
			//nolint:err113 // includes dynamic group version for debugging
			fluxInstanceGroupVersion.String(): fmt.Errorf("group %s unavailable", fluxInstanceGroupVersion.String()),
		}), nil
	})

	overrideFluxResourcesClientFactory(t, func(*rest.Config) (client.Client, error) {
		t.Fatalf("flux resource client should not be created when APIs unavailable")

		//nolint:nilnil // unreachable after t.Fatalf
		return nil, nil
	})

	cluster := &v1alpha1.Cluster{Spec: v1alpha1.Spec{SourceDirectory: "k8s"}}

	err := EnsureDefaultResources(context.Background(), "kubeconfig", cluster)
	require.Error(t, err)
	require.Contains(t, err.Error(), fluxInstanceGroupVersion.String())
	require.Contains(t, err.Error(), "timed out waiting for API")
}

func overrideFluxResourcesClientFactory(t *testing.T, factory func(*rest.Config) (client.Client, error)) {
	t.Helper()

	original := newFluxResourcesClient
	newFluxResourcesClient = factory

	t.Cleanup(func() {
		newFluxResourcesClient = original
	})
}

func overrideDiscoveryClientFactory(t *testing.T, factory func(*rest.Config) (discovery.DiscoveryInterface, error)) {
	t.Helper()

	original := newDiscoveryClient
	newDiscoveryClient = factory

	t.Cleanup(func() {
		newDiscoveryClient = original
	})
}

func overrideRESTConfigLoader(t *testing.T, loader func(string) (*rest.Config, error)) {
	t.Helper()

	original := loadRESTConfig
	loadRESTConfig = loader

	t.Cleanup(func() {
		loadRESTConfig = original
	})
}

func setFluxAPITimeouts(t *testing.T, timeout, interval time.Duration) {
	t.Helper()

	originalTimeout := fluxAPIAvailabilityTimeout
	originalInterval := fluxAPIAvailabilityPollInterval
	fluxAPIAvailabilityTimeout = timeout
	fluxAPIAvailabilityPollInterval = interval

	t.Cleanup(func() {
		fluxAPIAvailabilityTimeout = originalTimeout
		fluxAPIAvailabilityPollInterval = originalInterval
	})
}

type stubDiscoveryClient struct {
	*discoveryfake.FakeDiscovery

	responses map[string]error
	calls     int
}

func newStubDiscovery(responses map[string]error) *stubDiscoveryClient {
	if responses == nil {
		responses = map[string]error{}
	}

	return &stubDiscoveryClient{
		FakeDiscovery: &discoveryfake.FakeDiscovery{Fake: &clientgotesting.Fake{}},
		responses:     responses,
	}
}

func (s *stubDiscoveryClient) ServerResourcesForGroupVersion(gv string) (*metav1.APIResourceList, error) {
	s.calls++
	if err, ok := s.responses[gv]; ok && err != nil {
		return nil, err
	}

	return &metav1.APIResourceList{GroupVersion: gv}, nil
}
