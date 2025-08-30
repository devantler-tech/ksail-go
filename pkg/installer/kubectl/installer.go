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
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/yaml"
)

//go:embed assets/apply-set-crd.yaml
var applySetCRDYAML []byte

//go:embed assets/apply-set-cr.yaml
var applySetCRYAML []byte

// ErrCRDNameNotAccepted is returned when CRD names are not accepted.
var ErrCRDNameNotAccepted = errors.New("crd names not accepted")

// KubectlInstaller implements the installer.Installer interface for kubectl.
type KubectlInstaller struct {
	kubeconfig string
	context    string
	timeout    time.Duration
}

// NewKubectlInstaller creates a new kubectl installer instance.
func NewKubectlInstaller(kubeconfig, context string, timeout time.Duration) *KubectlInstaller {
	return &KubectlInstaller{
		kubeconfig: kubeconfig,
		context:    context,
		timeout:    timeout,
	}
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

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	gvr := schema.GroupVersionResource{Group: "k8s.devantler.tech", Version: "v1", Resource: "applysets"}
	_ = dynClient.Resource(gvr).Delete(ctx, "ksail", metav1.DeleteOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		GracePeriodSeconds:                    nil,
		Preconditions:                         nil,
		OrphanDependents:                      nil,
		PropagationPolicy:                     nil,
		DryRun:                                nil,
		IgnoreStoreReadErrorWithClusterBreakingPotential: func() *bool { b := false;

			return &b
		}(),
	}) // ignore errors (including NotFound)

	apiExtClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	_ = apiExtClient.ApiextensionsV1().
		CustomResourceDefinitions().
		Delete(ctx, "applysets.k8s.devantler.tech", metav1.DeleteOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "",
				APIVersion: "",
			},
			GracePeriodSeconds:                    nil,
			Preconditions:                         nil,
			OrphanDependents:                      nil,
			PropagationPolicy:                     nil,
			DryRun:                                nil,
			IgnoreStoreReadErrorWithClusterBreakingPotential: func() *bool { b := false;

				return &b
			}(),
		})

	return nil
}

// --- internals ---

// installCRD installs the ApplySet CRD.
func (b *KubectlInstaller) installCRD(restConfig *rest.Config) error {
	apiExtClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	const crdName = "applysets.k8s.devantler.tech"

	_, err = apiExtClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
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
	dynClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()

	gvr := schema.GroupVersionResource{Group: "k8s.devantler.tech", Version: "v1", Resource: "applysets"}

	const applySetName = "ksail"

	_, err = dynClient.Resource(gvr).Get(ctx, applySetName, metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		ResourceVersion: "",
	})
	if apierrors.IsNotFound(err) {
		err = b.applyApplySetCR(ctx, dynClient, gvr, applySetName)
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
		ExplicitPath:        kubeconfigPath,
		Precedence:          nil,
		MigrationRules:      nil,
		DoNotResolvePaths:   false,
		DefaultClientConfig: nil,
		WarnIfAllMissing:    false,
		Warner:              nil,
	}
}

func (b *KubectlInstaller) buildConfigOverrides() *clientcmd.ConfigOverrides {
	overrides := &clientcmd.ConfigOverrides{
		AuthInfo:        b.buildAuthInfo(),
		ClusterDefaults: b.buildCluster(),
		ClusterInfo:     b.buildCluster(),
		Context:         b.buildContext(),
		CurrentContext:  "",
		Timeout:         "",
	}
	if b.context != "" {
		overrides.CurrentContext = b.context
	}

	return overrides
}

func (b *KubectlInstaller) buildAuthInfo() api.AuthInfo {
	return api.AuthInfo{
		LocationOfOrigin:         "",
		ClientCertificate:        "",
		ClientCertificateData:    nil,
		ClientKey:                "",
		ClientKeyData:            nil,
		Token:                    "",
		TokenFile:                "",
		Impersonate:              "",
		ImpersonateUID:           "",
		ImpersonateGroups:        nil,
		ImpersonateUserExtra:     nil,
		Username:                 "",
		Password:                 "",
		AuthProvider:             nil,
		Exec:                     nil,
		Extensions:               nil,
	}
}

func (b *KubectlInstaller) buildCluster() api.Cluster {
	return api.Cluster{
		LocationOfOrigin:         "",
		Server:                   "",
		TLSServerName:            "",
		InsecureSkipTLSVerify:    false,
		CertificateAuthority:     "",
		CertificateAuthorityData: nil,
		ProxyURL:                 "",
		DisableCompression:       false,
		Extensions:               nil,
	}
}

func (b *KubectlInstaller) buildContext() api.Context {
	return api.Context{
		LocationOfOrigin: "",
		Cluster:          "",
		AuthInfo:         "",
		Namespace:        "",
		Extensions:       nil,
	}
}

// applyCRD creates the ApplySet CRD from embedded YAML.
func (b *KubectlInstaller) applyCRD(ctx context.Context, client *apiextensionsclient.Clientset) error {
	var crd apiextensionsv1.CustomResourceDefinition

	err := yaml.Unmarshal(applySetCRDYAML, &crd)
	if err != nil {
		return fmt.Errorf("failed to unmarshal CRD yaml: %w", err)
	}
	// Attempt create; if already exists attempt update (could race).
	_, err = client.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, &crd, metav1.CreateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun:          nil,
		FieldManager:    "",
		FieldValidation: "",
	})
	if apierrors.IsAlreadyExists(err) {
		existing, getErr := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd.Name, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "",
				APIVersion: "",
			},
			ResourceVersion: "",
		})
		if getErr != nil {
			return fmt.Errorf("failed to get existing CRD for update: %w", getErr)
		}

		crd.ResourceVersion = existing.ResourceVersion

		_, uerr := client.ApiextensionsV1().CustomResourceDefinitions().Update(ctx, &crd, metav1.UpdateOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "",
				APIVersion: "",
			},
			DryRun:          nil,
			FieldManager:    "",
			FieldValidation: "",
		})
		if uerr != nil {
			return fmt.Errorf("failed to update CRD: %w", uerr)
		}

		return nil
	}

	return fmt.Errorf("failed to create CRD: %w", err)
}

func (b *KubectlInstaller) waitForCRDEstablished(
	ctx context.Context,
	client *apiextensionsclient.Clientset,
	name string,
) error {
	// Poll every 500ms until Established=True or timeout
	const pollInterval = 500 * time.Millisecond

	err := wait.PollUntilContextTimeout(ctx, pollInterval, b.timeout, true,
		func(ctx context.Context) (bool, error) {
			crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "",
				APIVersion: "",
			},
			ResourceVersion: "",
		})
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
	dyn dynamic.Interface,
	gvr schema.GroupVersionResource,
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

	_, err = dyn.Resource(gvr).Create(ctx, &applySetObj, metav1.CreateOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "",
			APIVersion: "",
		},
		DryRun:          nil,
		FieldManager:    "",
		FieldValidation: "",
	})
	if apierrors.IsAlreadyExists(err) {
		existing, getErr := dyn.Resource(gvr).Get(ctx, name, metav1.GetOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "",
				APIVersion: "",
			},
			ResourceVersion: "",
		})
		if getErr != nil {
			return fmt.Errorf("failed to get existing ApplySet: %w", getErr)
		}

		applySetObj.SetResourceVersion(existing.GetResourceVersion())

		_, uerr := dyn.Resource(gvr).Update(ctx, &applySetObj, metav1.UpdateOptions{
			TypeMeta: metav1.TypeMeta{
				Kind:       "",
				APIVersion: "",
			},
			DryRun:          nil,
			FieldManager:    "",
			FieldValidation: "",
		})
		if uerr != nil {
			return fmt.Errorf("failed to update ApplySet: %w", uerr)
		}

		return nil
	}

	return fmt.Errorf("failed to create ApplySet CR: %w", err)
}
