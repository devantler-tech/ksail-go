package di

import "github.com/samber/do/v2"

// Module composition helpers.

// ProvideDependency returns a module that registers the supplied provider with the injector.
// This helper keeps service constructors in pkg/svc packages decoupled from the overall
// runtime wiring, enabling clean separation of concerns.
//
// Example usage:
//
//	module := ProvideDependency(func(i Injector) (*MyService, error) {
//		return NewMyService(), nil
//	})
func ProvideDependency[T any](provider func(Injector) (T, error)) Module {
	return func(injector Injector) error {
		do.Provide(injector, provider)

		return nil
	}
}

// ComposeModules bundles multiple modules into a single module for easier reuse.
// Nil modules are skipped, and execution stops immediately if any module returns an error.
//
// This is useful for grouping related dependencies or creating reusable module sets
// for different testing scenarios or command contexts.
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
