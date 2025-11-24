package di

import "github.com/samber/do/v2"

// ProvideDependency returns a module that registers the supplied provider with the injector.
// This helper keeps service constructors in pkg/svc packages decoupled from the overall runtime wiring.
func ProvideDependency[T any](provider func(Injector) (T, error)) Module {
	return func(injector Injector) error {
		do.Provide(injector, provider)

		return nil
	}
}

// ComposeModules bundles multiple modules into a single module for easier reuse.
// Nil modules are skipped and execution stops if any module returns an error.
func ComposeModules(modules ...Module) Module {
	return func(injector Injector) error {
		for _, module := range modules {
			if module == nil {
				continue
			}

			err := module(injector)
			if err != nil {
				return err
			}
		}

		return nil
	}
}
