// Package di exposes shared dependency injection helpers for KSail commands.
package di

import (
	"fmt"

	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
	"github.com/spf13/cobra"
)

// Module registers dependencies with the DI container.
type Module func(do.Injector) error

// Runtime owns shared dependency registration for Cobra commands.
type Runtime struct {
	baseModules []Module
}

// New constructs a Runtime with the provided base modules. Modules are applied
// in the order supplied when invoking commands.
func New(modules ...Module) *Runtime {
	return &Runtime{
		baseModules: append([]Module{}, modules...),
	}
}

// RunEWithRuntime returns a cobra RunE function that resolves dependencies using the provided runtime container.
// The handler is executed with the active command and resolved injector when the command runs.
func RunEWithRuntime(
	runtimeContainer *Runtime,
	handler func(cmd *cobra.Command, injector Injector) error,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return runtimeContainer.Invoke(func(injector Injector) error {
			return handler(cmd, injector)
		})
	}
}

// NewRuntime constructs the shared runtime container used by root command and tests.
func NewRuntime() *Runtime {
	return New(
		func(i Injector) error {
			do.Provide(i, func(Injector) (timer.Timer, error) {
				return timer.New(), nil
			})

			return nil
		},
		func(i Injector) error {
			do.Provide(i, func(Injector) (clusterprovisioner.Factory, error) {
				return clusterprovisioner.DefaultFactory{}, nil
			})

			return nil
		},
	)
}

// Injector is an alias for the underlying DI container implementation.
type Injector = do.Injector

// ResolveTimer retrieves the timer dependency from the injector with consistent error handling.
//
//nolint:ireturn,nolintlint // DI container exposes the timer interface.
func ResolveTimer(injector Injector) (timer.Timer, error) {
	tmr, err := do.Invoke[timer.Timer](injector)
	if err != nil {
		return nil, fmt.Errorf("resolve timer dependency: %w", err)
	}

	return tmr, nil
}

// ResolveClusterProvisionerFactory retrieves the cluster provisioner factory dependency.
//
//nolint:ireturn,nolintlint // DI container exposes the factory interface.
func ResolveClusterProvisionerFactory(
	injector Injector,
) (clusterprovisioner.Factory, error) {
	factory, err := do.Invoke[clusterprovisioner.Factory](injector)
	if err != nil {
		return nil, fmt.Errorf("resolve provisioner factory dependency: %w", err)
	}

	return factory, nil
}

// WithTimer decorates a handler to automatically resolve the timer dependency.
func WithTimer(
	handler func(cmd *cobra.Command, injector Injector, tmr timer.Timer) error,
) func(cmd *cobra.Command, injector Injector) error {
	return func(cmd *cobra.Command, injector Injector) error {
		tmr, err := ResolveTimer(injector)
		if err != nil {
			return err
		}

		return handler(cmd, injector, tmr)
	}
}

// Invoke builds a fresh injector, applies base and extra modules, and executes
// the provided function.
func (r *Runtime) Invoke(
	handler func(Injector) error,
	extraModules ...Module,
) error {
	injector := do.New()

	defer func() {
		_ = injector.Shutdown()
	}()

	modules := append([]Module{}, r.baseModules...)
	modules = append(modules, extraModules...)

	for _, module := range modules {
		if module == nil {
			continue
		}

		err := module(injector)
		if err != nil {
			return err
		}
	}

	return handler(injector)
}
