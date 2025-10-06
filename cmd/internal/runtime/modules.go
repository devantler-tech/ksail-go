package runtime

import (
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/samber/do/v2"
)

// ProvideProvisionerFactory registers the default provisioner factory implementation.
func ProvideProvisionerFactory() Module {
	return func(i do.Injector) error {
		do.Provide(i, func(do.Injector) (clusterprovisioner.Factory, error) {
			return clusterprovisioner.DefaultFactory{}, nil
		})

		return nil
	}
}

// ProvideTimer registers a module that creates a fresh timer for each invocation.
func ProvideTimer() Module {
	return func(i do.Injector) error {
		do.Provide(i, func(do.Injector) (timer.Timer, error) {
			return timer.New(), nil
		})

		return nil
	}
}

// WithProvisionerFactory overrides the provisioner factory, useful for tests.
func WithProvisionerFactory(factory clusterprovisioner.Factory) Module {
	return func(i do.Injector) error {
		do.Override(i, func(do.Injector) (clusterprovisioner.Factory, error) {
			return factory, nil
		})

		return nil
	}
}

// WithTimer injects a preconfigured timer implementation.
func WithTimer(t timer.Timer) Module {
	return func(i do.Injector) error {
		do.Override(i, func(do.Injector) (timer.Timer, error) {
			return t, nil
		})

		return nil
	}
}
