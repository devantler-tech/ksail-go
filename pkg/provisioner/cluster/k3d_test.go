package clusterprovisioner

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

var errBoom = errors.New("boom")

// --- helpers ---

// assertErrWrappedContains verifies an error exists, wraps a target error,
// and optionally contains a given substring in its message.
func assertErrWrappedContains(t *testing.T, got error, want error, contains string, ctx string) {
	t.Helper()

	if got == nil {
		t.Fatalf("%s: expected error, got nil", ctx)
	}

	if want != nil && !errors.Is(got, want) {
		t.Fatalf("%s: expected error to wrap %v, got %v", ctx, want, got)
	}

	if contains != "" && !strings.Contains(got.Error(), contains) {
		t.Fatalf("%s: expected error to contain %q, got %q", ctx, contains, got.Error())
	}
}

type k3dStub struct {
	runErr       error
	deleteErr    error
	getCluster   *types.Cluster
	getErr       error
	startErr     error
	stopErr      error
	listClusters []*types.Cluster
	listErr      error
	transformCfg *types.Cluster // not used; placeholder to carry name
	transformErr error
}

// withK3dStubs replaces k3d function variables with stubs for the duration of fn.
func withK3dStubs(t *testing.T, s k3dStub, fn func()) {
	t.Helper()
	// Save originals
	origRun := k3dClusterRun
	origDel := k3dClusterDelete
	origGet := k3dClusterGet
	origStart := k3dClusterStart
	origStop := k3dClusterStop
	origList := k3dClusterList
	origTransform := k3dTransformSimpleToClusterConfig

	// Install stubs
	k3dClusterRun = func(ctx context.Context, rt runtimes.Runtime, cfg *v1alpha5.ClusterConfig) error { // signature compatible with usage after transform
		return s.runErr
	}
	k3dClusterDelete = func(ctx context.Context, rt runtimes.Runtime, c *types.Cluster, opts types.ClusterDeleteOpts) error {
		return s.deleteErr
	}
	k3dClusterGet = func(ctx context.Context, rt runtimes.Runtime, c *types.Cluster) (*types.Cluster, error) {
		if s.getCluster != nil {
			return s.getCluster, nil
		}

		return nil, s.getErr
	}
	k3dClusterStart = func(ctx context.Context, rt runtimes.Runtime, c *types.Cluster, _ types.ClusterStartOpts) error {
		return s.startErr
	}
	k3dClusterStop = func(ctx context.Context, rt runtimes.Runtime, c *types.Cluster) error {
		return s.stopErr
	}
	k3dClusterList = func(ctx context.Context, rt runtimes.Runtime) ([]*types.Cluster, error) {
		if s.listClusters != nil {
			return s.listClusters, nil
		}

		return nil, s.listErr
	}
	k3dTransformSimpleToClusterConfig = func(ctx context.Context, rt runtimes.Runtime, simple v1alpha5.SimpleConfig, filename string) (*v1alpha5.ClusterConfig, error) {
		if s.transformErr != nil {
			return nil, s.transformErr
		}
		// Return minimal config; fields not used by tests
		return &v1alpha5.ClusterConfig{}, nil
	}

	// Restore after
	defer func() {
		k3dClusterRun = origRun
		k3dClusterDelete = origDel
		k3dClusterGet = origGet
		k3dClusterStart = origStart
		k3dClusterStop = origStop
		k3dClusterList = origList
		k3dTransformSimpleToClusterConfig = origTransform
	}()

	fn()
}

// --- tests ---

func TestCreate_Success(t *testing.T) {
	cases := []struct {
		name         string
		inputName    string
		expectedName string
	}{
		{name: "with name", inputName: "my-k3d", expectedName: "my-k3d"},
		{name: "without name uses cfg", inputName: "", expectedName: "cfg-name"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false

			withK3dStubs(t, k3dStub{}, func() {
				// Capture the name passed into the transform function
				k3dTransformSimpleToClusterConfig = func(ctx context.Context, rt runtimes.Runtime, simple v1alpha5.SimpleConfig, filename string) (*v1alpha5.ClusterConfig, error) {
					if simple.Name != tc.expectedName {
						return nil, fmt.Errorf("unexpected simple name: %s", simple.Name)
					}

					return &v1alpha5.ClusterConfig{}, nil
				}
				// Override run just to confirm it was called
				k3dClusterRun = func(ctx context.Context, rt runtimes.Runtime, cfg *v1alpha5.ClusterConfig) error {
					called = true

					return nil
				}

				simple := &v1alpha5.SimpleConfig{}
				simple.Name = "cfg-name"

				p := NewK3dClusterProvisioner(simple)

				err := p.Create(context.Background(), tc.inputName)
				if err != nil {
					t.Fatalf("Create() unexpected error: %v", err)
				}

				if !called {
					t.Fatalf("expected ClusterRun to be called")
				}
			})
		})
	}
}

func TestCreate_Error_TransformFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{transformErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		err := p.Create(context.Background(), "any")
		assertErrWrappedContains(t, err, errBoom, "failed to transform simple config to cluster config", "Create()")
	})
}

func TestCreate_Error_RunFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{runErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		err := p.Create(context.Background(), "any")
		assertErrWrappedContains(t, err, errBoom, "failed to run k3d cluster", "Create()")
	})
}

func TestDelete_Success(t *testing.T) {
	cases := []struct {
		name         string
		inputName    string
		expectedName string
	}{
		{name: "without name uses cfg", inputName: "", expectedName: "cfg-name"},
		{name: "with name", inputName: "custom", expectedName: "custom"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			called := false

			withK3dStubs(t, k3dStub{}, func() {
				k3dClusterDelete = func(ctx context.Context, rt runtimes.Runtime, c *types.Cluster, _ types.ClusterDeleteOpts) error {
					called = true

					if c == nil || c.Name != tc.expectedName {
						t.Fatalf("unexpected cluster name: %v", c)
					}

					return nil
				}

				simple := &v1alpha5.SimpleConfig{}
				simple.Name = "cfg-name"

				p := NewK3dClusterProvisioner(simple)

				err := p.Delete(context.Background(), tc.inputName)
				if err != nil {
					t.Fatalf("Delete() unexpected error: %v", err)
				}

				if !called {
					t.Fatalf("expected ClusterDelete to be called")
				}
			})
		})
	}
}

func TestDelete_Error_DeleteFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{deleteErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		err := p.Delete(context.Background(), "bad")
		assertErrWrappedContains(t, err, errBoom, "failed to delete k3d cluster \"bad\"", "Delete()")
	})
}

func TestStart_Success(t *testing.T) {
	withK3dStubs(t, k3dStub{getCluster: &types.Cluster{Name: "cfg-name"}}, func() {
		called := false
		k3dClusterStart = func(ctx context.Context, rt runtimes.Runtime, c *types.Cluster, _ types.ClusterStartOpts) error {
			called = true

			if c == nil || c.Name != "cfg-name" {
				t.Fatalf("unexpected cluster: %v", c)
			}

			return nil
		}
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"

		p := NewK3dClusterProvisioner(simple)

		err := p.Start(context.Background(), "")
		if err != nil {
			t.Fatalf("Start() unexpected error: %v", err)
		}

		if !called {
			t.Fatalf("expected ClusterStart to be called")
		}
	})
}

func TestStart_Error_GetFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{getErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		err := p.Start(context.Background(), "cfg-name")
		assertErrWrappedContains(t, err, errBoom, "failed to get k3d cluster \"cfg-name\"", "Start()")
	})
}

func TestStart_Error_StartFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{getCluster: &types.Cluster{Name: "cfg-name"}, startErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		err := p.Start(context.Background(), "cfg-name")
		assertErrWrappedContains(t, err, errBoom, "failed to start k3d cluster \"cfg-name\"", "Start()")
	})
}

func TestStop_Success(t *testing.T) {
	withK3dStubs(t, k3dStub{getCluster: &types.Cluster{Name: "my"}}, func() {
		called := false
		k3dClusterStop = func(ctx context.Context, rt runtimes.Runtime, c *types.Cluster) error {
			called = true

			if c == nil || c.Name != "my" {
				t.Fatalf("unexpected cluster: %v", c)
			}

			return nil
		}
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "my"

		p := NewK3dClusterProvisioner(simple)

		err := p.Stop(context.Background(), "")
		if err != nil {
			t.Fatalf("Stop() unexpected error: %v", err)
		}

		if !called {
			t.Fatalf("expected ClusterStop to be called")
		}
	})
}

func TestStop_Error_GetFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{getErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		err := p.Stop(context.Background(), "cfg-name")
		assertErrWrappedContains(t, err, errBoom, "failed to get k3d cluster \"cfg-name\"", "Stop()")
	})
}

func TestStop_Error_StopFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{getCluster: &types.Cluster{Name: "cfg-name"}, stopErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		err := p.Stop(context.Background(), "cfg-name")
		assertErrWrappedContains(t, err, errBoom, "failed to stop k3d cluster \"cfg-name\"", "Stop()")
	})
}

func TestList_Success(t *testing.T) {
	withK3dStubs(t, k3dStub{listClusters: []*types.Cluster{{Name: "a"}, {Name: "b"}}}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)

		got, err := p.List(context.Background())
		if err != nil {
			t.Fatalf("List() unexpected error: %v", err)
		}

		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Fatalf("unexpected list result: %v", got)
		}
	})
}

func TestList_Error_ListFailed(t *testing.T) {
	withK3dStubs(t, k3dStub{listErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		_, err := p.List(context.Background())
		assertErrWrappedContains(t, err, errBoom, "failed to list k3d clusters", "List()")
	})
}

func TestExists(t *testing.T) {
	withK3dStubs(t, k3dStub{listClusters: []*types.Cluster{{Name: "x"}, {Name: "cfg-name"}}}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)

		exists, err := p.Exists(context.Background(), "")
		if err != nil {
			t.Fatalf("Exists() unexpected error: %v", err)
		}

		if !exists {
			t.Fatalf("expected cluster to exist")
		}
	})

	withK3dStubs(t, k3dStub{listClusters: []*types.Cluster{{Name: "x"}, {Name: "y"}}}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)

		exists, err := p.Exists(context.Background(), "z")
		if err != nil {
			t.Fatalf("Exists() unexpected error: %v", err)
		}

		if exists {
			t.Fatalf("expected cluster to not exist")
		}
	})

	withK3dStubs(t, k3dStub{listErr: errBoom}, func() {
		simple := &v1alpha5.SimpleConfig{}
		simple.Name = "cfg-name"
		p := NewK3dClusterProvisioner(simple)
		_, err := p.Exists(context.Background(), "any")
		assertErrWrappedContains(t, err, errBoom, "failed to list k3d clusters", "Exists()")
	})
}
