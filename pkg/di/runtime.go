package di

import "github.com/samber/do/v2"

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

// Invoke builds a fresh injector, applies base and extra modules, and executes
// the provided function.
func (r *Runtime) Invoke(
	fn func(do.Injector) error,
	extraModules ...Module,
) error {
	injector := do.New()
	defer injector.Shutdown()

	modules := append([]Module{}, r.baseModules...)
	modules = append(modules, extraModules...)

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
