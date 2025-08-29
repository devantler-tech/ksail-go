// Package providerprovisioner provides an interface for managing providers.
package providerprovisioner

// ProviderProvisioner defines methods for managing providers.
type ProviderProvisioner interface {
	CheckReady() (bool, error)
}
