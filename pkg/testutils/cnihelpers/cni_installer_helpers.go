package cnihelpers

import (
	"context"
	"reflect"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/client/helm"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
)

// CNIInstaller defines the minimal interface needed for testing CNI installers.
type CNIInstaller interface {
	Install(ctx context.Context) error
	Uninstall(ctx context.Context) error
	WaitForReadiness(ctx context.Context) error
	SetWaitForReadinessFunc(waitFunc func(ctx context.Context) error)
	GetWaitFn() func(ctx context.Context) error
	SetWaitFn(fn func(ctx context.Context) error)
}

// InstallerScenario defines a test scenario for installer operations.
type InstallerScenario[T CNIInstaller] struct {
	Name       string
	Setup      func(*testing.T, *helm.MockInterface)
	ActionName string
	Action     func(context.Context, T) error
	WantErr    string
}

// RunInstallerScenarios runs a set of installer test scenarios.
func RunInstallerScenarios[T CNIInstaller](
	t *testing.T,
	scenarios []InstallerScenario[T],
	newInstaller func(*testing.T) (T, *helm.MockInterface),
) {
	t.Helper()

	for _, scenario := range scenarios {
		t.Run(scenario.Name, func(t *testing.T) {
			t.Parallel()

			installer, client := newInstaller(t)
			scenario.Setup(t, client)

			err := scenario.Action(context.Background(), installer)

			ExpectInstallerResult(t, err, scenario.WantErr, scenario.ActionName)
		})
	}
}

// TestSetWaitForReadinessFunc tests the SetWaitForReadinessFunc method for any CNI installer.
//
//nolint:tparallel // Helper is called from parallel tests; subtests run in parallel.
func TestSetWaitForReadinessFunc[T CNIInstaller](
	t *testing.T,
	newInstaller func(*testing.T) T,
) {
	t.Helper()

	t.Run("InvokesCustomFunction", func(t *testing.T) {
		t.Parallel()

		installer := newInstaller(t)
		called := false

		installer.SetWaitForReadinessFunc(func(context.Context) error {
			called = true

			return nil
		})

		testutils.ExpectNoError(
			t,
			installer.WaitForReadiness(context.Background()),
			"WaitForReadiness with custom func",
		)
		testutils.ExpectTrue(t, called, "custom wait function invocation")
	})

	t.Run("RestoresDefaultWhenNil", func(t *testing.T) {
		t.Parallel()

		installer := newInstaller(t)
		defaultFn := installer.GetWaitFn()
		testutils.ExpectNotNil(t, defaultFn, "default wait function")
		defaultPtr := reflect.ValueOf(defaultFn).Pointer()

		installer.SetWaitForReadinessFunc(func(context.Context) error { return nil })

		replacedPtr := reflect.ValueOf(installer.GetWaitFn()).Pointer()
		if replacedPtr == defaultPtr {
			t.Fatal("expected custom wait function to replace default")
		}

		installer.SetWaitForReadinessFunc(nil)
		restoredPtr := reflect.ValueOf(installer.GetWaitFn()).Pointer()
		ExpectEqual(
			t,
			restoredPtr,
			defaultPtr,
			"wait function pointer after restore",
		)
	})
}

// TestWaitForReadinessNoOpWhenUnset tests behavior when wait function is unset.
func TestWaitForReadinessNoOpWhenUnset[T CNIInstaller](
	t *testing.T,
	newInstaller func(*testing.T) T,
) {
	t.Helper()

	installer := newInstaller(t)
	installer.SetWaitFn(nil)

	err := installer.WaitForReadiness(context.Background())
	if err != nil {
		t.Fatalf("expected nil error when waitFn unset, got %v", err)
	}
}

// TestWaitForReadinessDetectsUnready tests detection of unready components.
// waitForReadiness is the function to test (typically a method that checks component readiness).
func TestWaitForReadinessDetectsUnready(
	t *testing.T,
	waitForReadiness func(context.Context) error,
) {
	t.Helper()

	err := waitForReadiness(context.Background())
	if err == nil {
		t.Fatal("expected readiness failure when components are unready")
	}

	if !containsSubstring(err.Error(), "not ready") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
