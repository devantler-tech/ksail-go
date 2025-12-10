package cluster //nolint:testpackage // Access helpers for white-box testing.

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	cmdhelpers "github.com/devantler-tech/ksail-go/pkg/cmd"
	registry "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/registry"
	testutils "github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/docker/docker/client"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var errUnexpectedRegistryInvocation = errors.New("unexpected registry service invocation")

const (
	kindNetworkName = "kind"
)

func TestExecuteLocalRegistryStage_SkipsWhenRegistryDisabled(t *testing.T) {
	t.Parallel()

	cmd, deps := newLocalRegistryCommandDeps(t)

	clusterCfg := v1alpha1.NewCluster()
	clusterCfg.Spec.LocalRegistry = v1alpha1.LocalRegistryDisabled

	neverCalled := withLocalRegistryServiceFactory(func(registry.Config) (registry.Service, error) {
		t.Fatalf("registry service should not be created when registry is disabled")

		return nil, errUnexpectedRegistryInvocation
	})

	err := executeLocalRegistryStage(
		cmd,
		clusterCfg,
		deps,
		nil,
		nil,
		localRegistryStageProvision,
		neverCalled,
	)
	if err != nil {
		t.Fatalf("expected nil error when registry disabled, got %v", err)
	}
}

func TestExecuteLocalRegistryStage_ProvisionCreatesAndStarts(t *testing.T) {
	t.Parallel()

	stub, cmd, deps, clusterCfg, option := newLocalRegistryTestEnv(t)
	clusterCfg.Spec.Options.LocalRegistry.HostPort = 5501
	clusterCfg.Spec.Distribution = v1alpha1.DistributionKind

	kindCfg := &kindv1alpha4.Cluster{Name: "kind-dev"}

	err := executeLocalRegistryStage(
		cmd,
		clusterCfg,
		deps,
		kindCfg,
		nil,
		localRegistryStageProvision,
		option,
	)
	if err != nil {
		t.Fatalf("expected provisioning to succeed, got %v", err)
	}

	if len(stub.createCalls) != 1 {
		t.Fatalf("expected 1 create call, got %d", len(stub.createCalls))
	}

	createOpts := stub.createCalls[0]
	if createOpts.Port != 5501 {
		t.Fatalf("expected port 5501, got %d", createOpts.Port)
	}

	if createOpts.Name != registry.LocalRegistryContainerName {
		t.Fatalf("unexpected registry name %q", createOpts.Name)
	}

	if len(stub.startCalls) != 1 {
		t.Fatalf("expected 1 start call, got %d", len(stub.startCalls))
	}

	if stub.startCalls[0].Name != createOpts.Name {
		t.Fatalf("start options should reference created registry")
	}
}

func TestExecuteLocalRegistryStage_ConnectAttachesNetwork(t *testing.T) {
	t.Parallel()

	stub, cmd, deps, clusterCfg, option := newLocalRegistryTestEnv(t)
	clusterCfg.Spec.Distribution = v1alpha1.DistributionK3d
	clusterCfg.Spec.Connection.Context = "dev"

	k3dCfg := &k3dv1alpha5.SimpleConfig{}
	k3dCfg.Name = "demo"

	err := executeLocalRegistryStage(
		cmd,
		clusterCfg,
		deps,
		nil,
		k3dCfg,
		localRegistryStageConnect,
		option,
	)
	if err != nil {
		t.Fatalf("expected connection to succeed, got %v", err)
	}

	if len(stub.startCalls) != 1 {
		t.Fatalf("expected 1 start call, got %d", len(stub.startCalls))
	}

	if stub.startCalls[0].NetworkName != "k3d-demo" {
		t.Fatalf("expected network k3d-demo, got %q", stub.startCalls[0].NetworkName)
	}
}

func TestCleanupLocalRegistry_DeletesWithVolumeFlag(t *testing.T) {
	t.Parallel()

	stub, cmd, deps, clusterCfg, option := newLocalRegistryTestEnv(t)
	tempDir := t.TempDir()
	kindConfigPath := filepath.Join(tempDir, "kind.yaml")

	err := os.WriteFile(
		kindConfigPath,
		[]byte("kind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\nname: local-kind\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to write kind config: %v", err)
	}

	cmd.SetContext(context.Background())

	clusterCfg.Spec.Distribution = v1alpha1.DistributionKind
	clusterCfg.Spec.DistributionConfig = kindConfigPath

	err = cleanupLocalRegistry(cmd, clusterCfg, deps, true, option)
	if err != nil {
		t.Fatalf("expected cleanup to succeed, got %v", err)
	}

	if len(stub.stopCalls) != 1 {
		t.Fatalf("expected 1 stop call, got %d", len(stub.stopCalls))
	}

	stopOpts := stub.stopCalls[0]
	if !stopOpts.DeleteVolume {
		t.Fatalf("expected delete volume to propagate")
	}

	if stopOpts.NetworkName != kindNetworkName {
		t.Fatalf("expected kind network, got %q", stopOpts.NetworkName)
	}
}

func newLocalRegistryTestEnv(
	t *testing.T,
) (
	*stubRegistryService,
	*cobra.Command,
	cmdhelpers.LifecycleDeps,
	*v1alpha1.Cluster,
	localRegistryOption,
) {
	t.Helper()

	stub := newStubRegistryService()

	withStubDockerInvoker(t)

	cmd, deps := newLocalRegistryCommandDeps(t)

	clusterCfg := v1alpha1.NewCluster()
	clusterCfg.Spec.LocalRegistry = v1alpha1.LocalRegistryEnabled

	option := withLocalRegistryServiceFactory(func(registry.Config) (registry.Service, error) {
		return stub, nil
	})

	return stub, cmd, deps, clusterCfg, option
}

type stubRegistryService struct {
	createCalls []registry.CreateOptions
	startCalls  []registry.StartOptions
	stopCalls   []registry.StopOptions
}

func newStubRegistryService() *stubRegistryService {
	return &stubRegistryService{}
}

func (s *stubRegistryService) Create(
	_ context.Context,
	opts registry.CreateOptions,
) (v1alpha1.OCIRegistry, error) {
	s.createCalls = append(s.createCalls, opts)

	return v1alpha1.NewOCIRegistry(), nil
}

func (s *stubRegistryService) Start(
	_ context.Context,
	opts registry.StartOptions,
) (v1alpha1.OCIRegistry, error) {
	s.startCalls = append(s.startCalls, opts)

	return v1alpha1.NewOCIRegistry(), nil
}

func (s *stubRegistryService) Stop(_ context.Context, opts registry.StopOptions) error {
	s.stopCalls = append(s.stopCalls, opts)

	return nil
}

func (s *stubRegistryService) Status(
	context.Context,
	registry.StatusOptions,
) (v1alpha1.OCIRegistry, error) {
	return v1alpha1.NewOCIRegistry(), nil
}

func withStubDockerInvoker(t *testing.T) {
	t.Helper()

	dockerClientInvokerMu.Lock()
	originalInvoker := dockerClientInvoker
	dockerClientInvoker = func(_ *cobra.Command, operation func(client.APIClient) error) error {
		return operation(nil)
	}
	dockerClientInvokerMu.Unlock()

	t.Cleanup(func() {
		dockerClientInvokerMu.Lock()
		dockerClientInvoker = originalInvoker
		dockerClientInvokerMu.Unlock()
	})
}

func newLocalRegistryCommandDeps(t *testing.T) (*cobra.Command, cmdhelpers.LifecycleDeps) {
	t.Helper()

	cmd, _ := testutils.NewCommand(t)
	cmd.SetContext(context.Background())

	deps := cmdhelpers.LifecycleDeps{Timer: &testutils.RecordingTimer{}}

	return cmd, deps
}
