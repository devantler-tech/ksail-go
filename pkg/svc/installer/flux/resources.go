package fluxinstaller

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	fluxclient "github.com/devantler-tech/ksail-go/pkg/client/flux"
	registry "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/registry"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultProjectName       = "ksail-workloads"
	defaultSourceDirectory   = "k8s"
	defaultArtifactTag       = "latest"
	defaultOCIRepositoryName = fluxclient.DefaultNamespace
	fluxIntervalFallback     = time.Minute
	fluxDistributionVersion  = "2.x"
	fluxDistributionRegistry = "ghcr.io/fluxcd"
	fluxDistributionArtifact = "oci://ghcr.io/controlplaneio-fluxcd/flux-operator-manifests:latest"
)

var (
	fluxAPIAvailabilityTimeout      = 2 * time.Minute
	fluxAPIAvailabilityPollInterval = 2 * time.Second
)

var (
	errInvalidClusterConfig = errors.New("cluster configuration is required")

	loadRESTConfig = buildRESTConfig

	newFluxResourcesClient = func(restConfig *rest.Config) (client.Client, error) {
		scheme := runtime.NewScheme()

		if err := addFluxInstanceToScheme(scheme); err != nil {
			return nil, fmt.Errorf("failed to add flux instance scheme: %w", err)
		}

		if err := sourcev1.AddToScheme(scheme); err != nil {
			return nil, fmt.Errorf("failed to add flux source scheme: %w", err)
		}

		fluxClient, err := client.New(restConfig, client.Options{Scheme: scheme})
		if err != nil {
			return nil, fmt.Errorf("failed to create flux resource client: %w", err)
		}

		return fluxClient, nil
	}

	newDiscoveryClient = func(restConfig *rest.Config) (discovery.DiscoveryInterface, error) {
		return discovery.NewDiscoveryClientForConfig(restConfig)
	}
)

// EnsureDefaultResources configures a default FluxInstance so the operator can
// bootstrap controllers and sync from the local OCI registry.
func EnsureDefaultResources(
	ctx context.Context,
	kubeconfig string,
	clusterCfg *v1alpha1.Cluster,
) error {
	if clusterCfg == nil {
		return errInvalidClusterConfig
	}

	if ctx == nil {
		ctx = context.Background()
	}

	restConfig, err := loadRESTConfig(kubeconfig)
	if err != nil {
		return err
	}

	if err := waitForGroupVersion(ctx, restConfig, fluxInstanceGroupVersion); err != nil {
		return err
	}

	fluxInstance, err := buildFluxInstance(clusterCfg)
	if err != nil {
		return err
	}

	fluxClient, err := newFluxResourcesClient(restConfig)
	if err != nil {
		return err
	}

	if err := upsertFluxResource(ctx, fluxClient, fluxInstance); err != nil {
		return err
	}

	if err := waitForGroupVersion(ctx, restConfig, sourcev1.GroupVersion); err != nil {
		return err
	}

	if clusterCfg.Spec.LocalRegistry == v1alpha1.LocalRegistryEnabled {
		return ensureLocalOCIRepositoryInsecure(ctx, fluxClient)
	}

	return nil
}

func buildFluxInstance(clusterCfg *v1alpha1.Cluster) (*FluxInstance, error) {
	interval := clusterCfg.Spec.Options.Flux.Interval.Duration
	if interval <= 0 {
		interval = fluxIntervalFallback
	}

	hostPort := clusterCfg.Spec.Options.LocalRegistry.HostPort
	if hostPort == 0 {
		hostPort = v1alpha1.DefaultLocalRegistryPort
	}

	sourceDir := strings.TrimSpace(clusterCfg.Spec.SourceDirectory)
	if sourceDir == "" {
		sourceDir = defaultSourceDirectory
	}

	projectName := sanitizeFluxName(sourceDir, defaultProjectName)
	repoHost := registry.LocalRegistryClusterHost
	repoPort := registry.DefaultRegistryPort

	if clusterCfg.Spec.LocalRegistry != v1alpha1.LocalRegistryEnabled {
		repoHost = registry.DefaultEndpointHost
		repoPort = int(hostPort)
	}

	repoURL := fmt.Sprintf("oci://%s:%d/%s", repoHost, repoPort, projectName)
	normalizedPath := normalizeFluxPath(sourceDir)
	intervalPtr := &metav1.Duration{Duration: interval}

	return &FluxInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fluxInstanceDefaultName,
			Namespace: fluxclient.DefaultNamespace,
		},
		Spec: FluxInstanceSpec{
			Distribution: Distribution{
				Version:  fluxDistributionVersion,
				Registry: fluxDistributionRegistry,
				Artifact: fluxDistributionArtifact,
			},
			Sync: &Sync{
				Kind:     fluxOCIRepositoryKind,
				URL:      repoURL,
				Ref:      defaultArtifactTag,
				Path:     normalizedPath,
				Provider: "generic",
				Interval: intervalPtr,
			},
		},
	}, nil
}

func upsertFluxResource(
	ctx context.Context,
	fluxClient client.Client,
	obj client.Object,
) error {
	key := client.ObjectKeyFromObject(obj)

	switch desired := obj.(type) {
	case *FluxInstance:
		existing := &FluxInstance{}
		if err := fluxClient.Get(ctx, key, existing); err != nil {
			if apierrors.IsNotFound(err) {
				return fluxClient.Create(ctx, desired)
			}

			return fmt.Errorf("failed to get FluxInstance %s/%s: %w", key.Namespace, key.Name, err)
		}

		existing.Spec = desired.Spec

		if err := fluxClient.Update(ctx, existing); err != nil {
			return fmt.Errorf("failed to update FluxInstance %s/%s: %w", key.Namespace, key.Name, err)
		}

		return nil
	default:
		return fmt.Errorf("unsupported Flux resource type %T", obj)
	}
}

func ensureLocalOCIRepositoryInsecure(ctx context.Context, fluxClient client.Client) error {
	key := client.ObjectKey{Name: defaultOCIRepositoryName, Namespace: fluxclient.DefaultNamespace}
	waitCtx, cancel := context.WithTimeout(ctx, fluxAPIAvailabilityTimeout)
	defer cancel()

	ticker := time.NewTicker(fluxAPIAvailabilityPollInterval)
	defer ticker.Stop()

	for {
		repo := &sourcev1.OCIRepository{}
		err := fluxClient.Get(ctx, key, repo)
		switch {
		case err == nil:
			if repo.Spec.Insecure {
				return nil
			}

			repo.Spec.Insecure = true
			if err := fluxClient.Update(ctx, repo); err != nil {
				return fmt.Errorf("failed to update OCIRepository %s/%s: %w", key.Namespace, key.Name, err)
			}

			return nil
		case apierrors.IsNotFound(err):
			select {
			case <-waitCtx.Done():
				return fmt.Errorf("timed out waiting for OCIRepository %s/%s", key.Namespace, key.Name)
			case <-ticker.C:
			}
		default:
			return fmt.Errorf("failed to get OCIRepository %s/%s: %w", key.Namespace, key.Name, err)
		}
	}
}

func sanitizeFluxName(value, fallback string) string {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if trimmed == "" {
		trimmed = fallback
	}

	var builder strings.Builder
	previousHyphen := false

	for _, r := range trimmed {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			previousHyphen = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			previousHyphen = false
		default:
			if !previousHyphen {
				builder.WriteRune('-')
				previousHyphen = true
			}
		}
	}

	sanitized := strings.Trim(builder.String(), "-")
	if sanitized == "" {
		sanitized = fallback
	}

	if len(sanitized) > validation.DNS1123LabelMaxLength {
		sanitized = sanitized[:validation.DNS1123LabelMaxLength]
		sanitized = strings.Trim(sanitized, "-")
	}

	if sanitized == "" {
		sanitized = fallback
	}

	if len(validation.IsDNS1123Label(sanitized)) == 0 {
		return sanitized
	}

	return fallback
}

func normalizeFluxPath(path string) string {
	// Flux expects paths to be relative to the root of the unpacked artifact.
	return "./"
}

func waitForGroupVersion(ctx context.Context, restConfig *rest.Config, groupVersion schema.GroupVersion) error {
	discoveryClient, err := newDiscoveryClient(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create discovery client: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, fluxAPIAvailabilityTimeout)
	defer cancel()

	ticker := time.NewTicker(fluxAPIAvailabilityPollInterval)
	defer ticker.Stop()

	var lastErr error
	for {
		if _, err := discoveryClient.ServerResourcesForGroupVersion(groupVersion.String()); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-waitCtx.Done():
			if lastErr == nil {
				lastErr = waitCtx.Err()
			}
			return fmt.Errorf("timed out waiting for API %s: %w", groupVersion.String(), lastErr)
		case <-ticker.C:
		}
	}
}

func buildRESTConfig(kubeconfig string) (*rest.Config, error) {
	if strings.TrimSpace(kubeconfig) == "" {
		return nil, errors.New("kubeconfig path is required")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig %s: %w", kubeconfig, err)
	}

	return restConfig, nil
}
