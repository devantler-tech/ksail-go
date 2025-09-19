// Package validator provides interfaces for validating configurations.
package validator

// Validator is an interface for validating configurations.
type Validator interface {
	Validate() error
}
