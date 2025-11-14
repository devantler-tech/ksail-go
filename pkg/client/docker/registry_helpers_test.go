package docker_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

var (
	errInspectFailed     = errors.New("inspect failed")
	errNetworkDisconnect = errors.New("network disconnect failed")
)

// setupRegistryWithContainer creates a common test setup with a running registry container.
func setupRegistryWithContainer(
	t *testing.T,
) (*docker.MockAPIClient, *docker.RegistryManager, context.Context) {
	t.Helper()
	mockClient, manager, ctx := setupTestRegistryManager(t)
	registry := mockRegistryContainer("registry-id", "docker.io", "test-cluster", "running")
	mockContainerListOnce(ctx, mockClient, []container.Summary{registry})

	return mockClient, manager, ctx
}

// TestRegistryManager_CreateRegistry_ContainerListError tests error handling when container list fails.
func TestRegistryManager_CreateRegistry_ContainerListError(t *testing.T) {
	t.Parallel()

	mockClient, manager, ctx := setupTestRegistryManager(t)
	config := newTestRegistryConfig()

	mockContainerListError(ctx, mockClient)

	err := manager.CreateRegistry(ctx, config)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to check if registry exists")
}

// TestListRegistries_WithEmptyLabel tests listing when label is empty but name exists.
func TestListRegistries_WithEmptyLabel(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupTestRegistryManager(t)
	mockContainerListOnce(ctx, mockClient, []container.Summary{
		{
			ID:     "registry1",
			Names:  []string{"/docker.io"},
			Labels: map[string]string{},
			Image:  docker.RegistryImageName,
		},
	})
	registries, err := manager.ListRegistries(ctx)
	require.NoError(t, err)
	require.Len(t, registries, 1)
	require.Contains(t, registries, "docker.io")
}

// TestListRegistries_WithDuplicateNames tests deduplication of registry names.
func TestListRegistries_WithDuplicateNames(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupTestRegistryManager(t)
	mockContainerListOnce(ctx, mockClient, []container.Summary{
		{
			ID:     "registry1",
			Names:  []string{"/docker.io"},
			Labels: map[string]string{docker.RegistryLabelKey: "docker.io"},
			Image:  docker.RegistryImageName,
		},
		{
			ID:     "registry2",
			Names:  []string{"/docker.io"},
			Labels: map[string]string{docker.RegistryLabelKey: "docker.io"},
			Image:  docker.RegistryImageName,
		},
	})
	registries, err := manager.ListRegistries(ctx)
	require.NoError(t, err)
	require.Len(t, registries, 1, "should deduplicate registry names")
	require.Contains(t, registries, "docker.io")
}

// TestListRegistries_SkipsEmptyNames tests that registries with no name are skipped.
func TestListRegistries_SkipsEmptyNames(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupTestRegistryManager(t)
	mockContainerListOnce(ctx, mockClient, []container.Summary{
		{
			ID:     "registry1",
			Names:  []string{},
			Labels: map[string]string{},
			Image:  docker.RegistryImageName,
		},
		{
			ID:     "registry2",
			Names:  []string{"/valid-registry"},
			Labels: map[string]string{docker.RegistryLabelKey: "valid-registry"},
			Image:  docker.RegistryImageName,
		},
	})
	registries, err := manager.ListRegistries(ctx)
	require.NoError(t, err)
	require.Len(t, registries, 1)
	require.Contains(t, registries, "valid-registry")
}

// TestDeleteRegistry_ContainerInspectError tests error handling when container inspect fails.
func TestDeleteRegistry_ContainerInspectError(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupRegistryWithContainer(t)
	mockClient.EXPECT().
		ContainerInspect(ctx, "registry-id").
		Return(container.InspectResponse{}, errInspectFailed).
		Once()
	err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to inspect registry container")
}

// TestDeleteRegistry_NetworkDisconnectError tests error handling when network disconnect fails.
func TestDeleteRegistry_NetworkDisconnectError(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupRegistryWithContainer(t)
	mockClient.EXPECT().
		ContainerInspect(ctx, "registry-id").
		Return(newInspectResponse("k3d-test"), nil).
		Once()
	mockClient.EXPECT().
		NetworkDisconnect(ctx, "k3d-test", "registry-id", true).
		Return(errNetworkDisconnect).
		Once()
	err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false, "k3d-test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to disconnect registry")
}

// TestDeleteRegistry_EmptyNetworkName tests deletion when network name is empty.
func TestDeleteRegistry_EmptyNetworkName(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupRegistryWithContainer(t)
	mockClient.EXPECT().
		ContainerInspect(ctx, "registry-id").
		Return(newInspectResponse(), nil).
		Once()
	mockContainerStopRemove(ctx, mockClient, "registry-id")
	err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false, "")
	require.NoError(t, err)
}

// TestRegistryAttachedToOtherClusters_EmptyNetworks tests when network settings is nil.
func TestRegistryAttachedToOtherClusters_EmptyNetworks(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupRegistryWithContainer(t)
	inspectResp := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{},
		NetworkSettings:   nil,
	}
	mockClient.EXPECT().
		ContainerInspect(ctx, "registry-id").
		Return(inspectResp, nil).
		Once()
	mockContainerStopRemove(ctx, mockClient, "registry-id")
	err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false, "")
	require.NoError(t, err)
}

// TestRegistryAttachedToOtherClusters_IgnoredNetworkMatching tests case-insensitive ignored network matching.
func TestRegistryAttachedToOtherClusters_IgnoredNetworkMatching(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupRegistryWithContainer(t)
	mockClient.EXPECT().
		ContainerInspect(ctx, "registry-id").
		Return(newInspectResponse("K3D-test"), nil).
		Once()
	mockClient.EXPECT().
		NetworkDisconnect(ctx, "k3d-test", "registry-id", true).
		Return(nil).
		Once()
	mockClient.EXPECT().
		ContainerInspect(ctx, "registry-id").
		Return(newInspectResponse(), nil).
		Once()
	mockContainerStopRemove(ctx, mockClient, "registry-id")
	err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false, "k3d-test")
	require.NoError(t, err)
}

// TestRegistryAttachedToOtherClusters_NonClusterNetworks tests that non-cluster networks don't prevent deletion.
func TestRegistryAttachedToOtherClusters_NonClusterNetworks(t *testing.T) {
	t.Parallel()
	mockClient, manager, ctx := setupRegistryWithContainer(t)
	mockClient.EXPECT().
		ContainerInspect(ctx, "registry-id").
		Return(newInspectResponse("bridge", "host"), nil).
		Once()
	mockContainerStopRemove(ctx, mockClient, "registry-id")
	err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false, "")
	require.NoError(t, err)
}

// TestDeleteRegistry_ListError tests error handling when container list fails.
func TestDeleteRegistry_ListError(t *testing.T) {
	t.Parallel()

	mockClient, manager, ctx := setupTestRegistryManager(t)

	mockContainerListError(ctx, mockClient)

	err := manager.DeleteRegistry(ctx, "docker.io", "test-cluster", false, "")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list registry containers")
}

// TestGetRegistryPort_ListError tests error handling when list fails.
func TestGetRegistryPort_ListError(t *testing.T) {
	t.Parallel()

	mockClient, manager, ctx := setupTestRegistryManager(t)

	mockContainerListError(ctx, mockClient)

	_, err := manager.GetRegistryPort(ctx, "docker.io")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to list registry containers")
}
