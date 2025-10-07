package di

import (
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

func TestNewCopiesModules(t *testing.T) {
	t.Parallel()

	called := 0
	mod := func(do.Injector) error {
		called++
		return nil
	}

	runtime := New(mod)
	if len(runtime.baseModules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(runtime.baseModules))
	}

	runtime.baseModules[0] = nil

	if runtime2 := New(mod); runtime2.baseModules[0] == nil {
		t.Fatal("expected New to copy supplied modules")
	}
}

func TestRunEWithRuntimeInvokesHandler(t *testing.T) {
	t.Parallel()

	runtime := New(func(injector Injector) error {
		do.Provide(injector, func(do.Injector) (string, error) { return "value", nil })
		return nil
	})

	handled := false
	handler := func(_ *cobra.Command, injector Injector) error {
		val, err := do.Invoke[string](injector)
		if err != nil {
			t.Fatalf("unexpected invoke error: %v", err)
		}
		if val != "value" {
			t.Fatalf("expected value, got %s", val)
		}
		handled = true
		return nil
	}

	cmd := &cobra.Command{}
	runE := RunEWithRuntime(runtime, handler)
	if err := runE(cmd, nil); err != nil {
		t.Fatalf("runE returned error: %v", err)
	}
	if !handled {
		t.Fatal("expected handler to be called")
	}
}

func TestNewRuntimeRegistersDefaults(t *testing.T) {
	t.Parallel()

	runtime := NewRuntime()
	err := runtime.Invoke(func(injector Injector) error {
		if _, err := ResolveTimer(injector); err != nil {
			t.Fatalf("resolve timer: %v", err)
		}
		if _, err := ResolveClusterProvisionerFactory(injector); err != nil {
			t.Fatalf("resolve factory: %v", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("invoke failed: %v", err)
	}
}

func TestResolveTimerError(t *testing.T) {
	t.Parallel()

	injector := do.New()
	defer injector.Shutdown()

	if _, err := ResolveTimer(injector); err == nil {
		t.Fatal("expected error when timer not registered")
	}
}

func TestResolveClusterProvisionerFactoryError(t *testing.T) {
	t.Parallel()

	injector := do.New()
	defer injector.Shutdown()

	if _, err := ResolveClusterProvisionerFactory(injector); err == nil {
		t.Fatal("expected error when factory not registered")
	}
}

func TestWithTimerSuccess(t *testing.T) {
	t.Parallel()

	called := false
	runtime := New(func(injector Injector) error {
		do.Provide(injector, func(do.Injector) (timer.Timer, error) {
			return timer.New(), nil
		})
		return nil
	})

	wrapped := WithTimer(func(_ *cobra.Command, _ Injector, tmr timer.Timer) error {
		if tmr == nil {
			t.Fatal("timer should not be nil")
		}
		called = true
		return nil
	})

	err := runtime.Invoke(func(injector Injector) error {
		return wrapped(&cobra.Command{}, injector)
	})
	if err != nil {
		t.Fatalf("wrapped handler returned error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
}

func TestWithTimerResolveErrorPropagates(t *testing.T) {
	t.Parallel()

	runtime := New()
	wrapped := WithTimer(func(_ *cobra.Command, _ Injector, _ timer.Timer) error { return nil })
	err := runtime.Invoke(func(injector Injector) error {
		return wrapped(&cobra.Command{}, injector)
	})
	if err == nil {
		t.Fatal("expected error when timer resolution fails")
	}
}

func TestRuntimeInvokeAppliesModulesAndExtra(t *testing.T) {
	t.Parallel()

	order := make([]string, 0, 2)

	base := func(do.Injector) error {
		order = append(order, "base")
		return nil
	}
	extra := func(do.Injector) error {
		order = append(order, "extra")
		return nil
	}

	runtime := New(base)

	err := runtime.Invoke(func(do.Injector) error {
		if len(order) != 2 {
			t.Fatalf("expected modules to run, order: %v", order)
		}
		if order[0] != "base" || order[1] != "extra" {
			t.Fatalf("unexpected module order: %v", order)
		}
		return nil
	}, extra)
	if err != nil {
		t.Fatalf("invoke returned error: %v", err)
	}
}

func TestRuntimeInvokeNilModuleIgnored(t *testing.T) {
	t.Parallel()

	runtime := New(nil)

	errSentinel := errors.New("failure")
	err := runtime.Invoke(
		func(do.Injector) error { return nil },
		func(do.Injector) error { return errSentinel },
	)
	if !errors.Is(err, errSentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
