package di

import (
	"context"

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

// Invoke builds a fresh injector for the command, applies base and extra
// modules, injects the command/context, and executes the provided function.
func (r *Runtime) Invoke(
	cmd *cobra.Command,
	fn func(do.Injector) error,
	extraModules ...Module,
) error {
	injector := do.New()
	defer injector.Shutdown()

	modules := append([]Module{}, r.baseModules...)
	modules = append(modules, extraModules...)
	modules = append(modules, provideCommand(cmd))

	for _, module := range modules {
		if module == nil {
			continue
		}

		if err := module(injector); err != nil {
			return err
		}
	}

	return fn(injector)
}

func provideCommand(cmd *cobra.Command) Module {
	return func(i do.Injector) error {
		do.ProvideValue(i, cmd)
		do.ProvideValue(i, commandContext(cmd))

		return nil
	}
}

func commandContext(cmd *cobra.Command) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}

	return context.Background()
}
