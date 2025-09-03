// Package kubectlinstaller provides a kubectl installer implementation.
package kubectlinstaller

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	pathutils "github.com/devantler-tech/ksail-go/internal/utils/path"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

//go:embed assets/apply-set-crd.yaml
var applySetCRDYAML []byte

//go:embed assets/apply-set-cr.yaml
var applySetCRYAML []byte

// boolPtr returns a pointer to the given boolean value.
func boolPtr(b bool) *bool {
	return &b
}

// createDefaultDeleteOptions creates a metav1.DeleteOptions with minimal necessary fields.
func createDefaultDeleteOptions() metav1.DeleteOptions {
	return metav1.DeleteOptions{
		IgnoreStoreReadErrorWithClusterBreakingPotential: boolPtr(false),
	}
}

// ErrCRDNameNotAccepted is returned when CRD names are not accepted.
var ErrCRDNameNotAccepted = errors.New("crd names not accepted")

// KubectlInstaller implements the installer.Installer interface for kubectl.
type KubectlInstaller struct {
	kubeconfig    string
	context       string
	timeout       time.Duration
	clientFactory ClientFactoryInterface
}

// NewKubectlInstaller creates a new kubectl installer instance.
func NewKubectlInstaller(kubeconfig, context string, timeout time.Duration, clientFactory ClientFactoryInterface) *KubectlInstaller {
	return &KubectlInstaller{
		kubeconfig:    kubeconfig,
		context:       context,
		timeout:       timeout,
		clientFactory: clientFactory,
	}
}

// NewKubectlInstallerWithFactory creates a new kubectl installer instance with a custom client factory.
// Deprecated: Use NewKubectlInstaller instead.
func NewKubectlInstallerWithFactory(kubeconfig, context string, timeout time.Duration, clientFactory ClientFactoryInterface) *KubectlInstaller {
	return NewKubectlInstaller(kubeconfig, context, timeout, clientFactory)
}

// Install ensures the ApplySet CRD and its parent CR exist.
func (b *KubectlInstaller) Install() error {
	restConfigWrapper, err := b.buildRESTConfig()
	if err != nil {
		return err
	}

	err = b.installCRD(restConfigWrapper)
	if err != nil {
		return err
	}

	err = b.installApplySetCR(restConfigWrapper)
	if err != nil {
		return err
	}

	return nil
}

// Uninstall deletes the ApplySet CR then its CRD.
func (b *KubectlInstaller) Uninstall() error {
	config, err := b.buildRESTConfig()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	gvr := schema.GroupVersionResource{Group: "k8s.devantler.tech", Version: "v1", Resource: "applysets"}
	
	dynClient, err := b.clientFactory.CreateDynamicClient(config, gvr)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	_ = dynClient.Delete(ctx, "ksail", createDefaultDeleteOptions()) // ignore errors (including NotFound)

	apiExtClient, err := b.clientFactory.CreateAPIExtensionsClient(config)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	_ = apiExtClient.Delete(ctx, "applysets.k8s.devantler.tech", createDefaultDeleteOptions())

	return nil
}

// --- internals ---

// installCRD installs the ApplySet CRD.
func (b *KubectlInstaller) installCRD(restConfig *rest.Config) error {
	apiExtClient, err := b.clientFactory.CreateAPIExtensionsClient(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	const crdName = "applysets.k8s.devantler.tech"

	_, err = apiExtClient.Get(ctx, crdName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		err = b.applyCRD(ctx, apiExtClient)
		if err != nil {
			return err
		}

		err = b.waitForCRDEstablished(ctx, apiExtClient, crdName)
		if err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("failed to check CRD existence: %w", err)
	}

	return nil
}

// installApplySetCR installs the ApplySet custom resource.
func (b *KubectlInstaller) installApplySetCR(restConfig *rest.Config) error {
	gvr := schema.GroupVersionResource{Group: "k8s.devantler.tech", Version: "v1", Resource: "applysets"}
	
	dynClient, err := b.clientFactory.CreateDynamicClient(restConfig, gvr)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	const applySetName = "ksail"

	_, err = dynClient.Get(ctx, applySetName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		err = b.applyApplySetCR(ctx, dynClient, applySetName)
		if err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("failed to get ApplySet CR: %w", err)
	}

	return nil
}

func (b *KubectlInstaller) buildRESTConfig() (*rest.Config, error) {
	kubeconfigPath, _ := pathutils.ExpandHomePath(b.kubeconfig)
	rules := b.buildClientConfigLoadingRules(kubeconfigPath)
	overrides := b.buildConfigOverrides()

	clientCfg := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)

	restConfig, err := clientCfg.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build rest config: %w", err)
	}

	return restConfig, nil
}

func (b *KubectlInstaller) buildClientConfigLoadingRules(kubeconfigPath string) *clientcmd.ClientConfigLoadingRules {
	return &clientcmd.ClientConfigLoadingRules{
		ExplicitPath: kubeconfigPath,
	}
}

func (b *KubectlInstaller) buildConfigOverrides() *clientcmd.ConfigOverrides {
	overrides := &clientcmd.ConfigOverrides{
		AuthInfo:     api.AuthInfo{},
		ClusterInfo:  api.Cluster{},
		Context:      api.Context{},
	}
	if b.context != "" {
		overrides.CurrentContext = b.context
	}

	return overrides
}

// applyCRD creates the ApplySet CRD from embedded YAML.
func (b *KubectlInstaller) applyCRD(ctx context.Context, client APIExtensionsClientInterface) error {
	var crd apiextensionsv1.CustomResourceDefinition

	err := yaml.Unmarshal(applySetCRDYAML, &crd)
	if err != nil {
		return fmt.Errorf("failed to unmarshal CRD yaml: %w", err)
	}
	// Attempt create; if already exists attempt update (could race).
	_, err = client.Create(ctx, &crd, metav1.CreateOptions{})
	if err == nil {
		return nil
	}
	if apierrors.IsAlreadyExists(err) {
		existing, getErr := client.Get(ctx, crd.Name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get existing CRD for update: %w", getErr)
		}

		crd.ResourceVersion = existing.ResourceVersion

		_, uerr := client.Update(ctx, &crd, metav1.UpdateOptions{})
		if uerr != nil {
			return fmt.Errorf("failed to update CRD: %w", uerr)
		}

		return nil
	}

	return fmt.Errorf("failed to create CRD: %w", err)
}

func (b *KubectlInstaller) waitForCRDEstablished(
	ctx context.Context,
	client APIExtensionsClientInterface,
	name string,
) error {
	// Poll every 500ms until Established=True or timeout
	const pollInterval = 500 * time.Millisecond

	err := wait.PollUntilContextTimeout(ctx, pollInterval, b.timeout, true,
		func(ctx context.Context) (bool, error) {
			crd, err := client.Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					return false, nil
				}

				return false, fmt.Errorf("failed to get CRD: %w", err)
			}

			for _, cond := range crd.Status.Conditions {
				if cond.Type == apiextensionsv1.Established && cond.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}

				if cond.Type == apiextensionsv1.NamesAccepted &&
					cond.Status == apiextensionsv1.ConditionFalse &&
					cond.Reason == "MultipleNamesNotAllowed" {
					return false, fmt.Errorf("%w: %s", ErrCRDNameNotAccepted, cond.Message)
				}
			}

			return false, nil
		})
	if err != nil {
		return fmt.Errorf("failed to wait for CRD to be established: %w", err)
	}

	return nil
}

func (b *KubectlInstaller) applyApplySetCR(
	ctx context.Context,
	dyn DynamicClientInterface,
	name string,
) error {
	var applySetObj unstructured.Unstructured

	err := yaml.Unmarshal(applySetCRYAML, &applySetObj.Object)
	if err != nil {
		return fmt.Errorf("failed to unmarshal ApplySet CR yaml: %w", err)
	}
	// Ensure GVK since yaml->map won't set it.
	applySetObj.SetGroupVersionKind(schema.GroupVersionKind{Group: "k8s.devantler.tech", Version: "v1", Kind: "ApplySet"})
	applySetObj.SetName(name)

	_, err = dyn.Create(ctx, &applySetObj, metav1.CreateOptions{})
	if err == nil {
		return nil
	}
	if apierrors.IsAlreadyExists(err) {
		existing, getErr := dyn.Get(ctx, name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get existing ApplySet: %w", getErr)
		}

		applySetObj.SetResourceVersion(existing.GetResourceVersion())

		_, uerr := dyn.Update(ctx, &applySetObj, metav1.UpdateOptions{})
		if uerr != nil {
			return fmt.Errorf("failed to update ApplySet: %w", uerr)
		}

		return nil
	}

	return fmt.Errorf("failed to create ApplySet CR: %w", err)
}
