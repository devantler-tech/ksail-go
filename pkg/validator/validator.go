package validators

// Validator is an interface for validating configurations.
type Validator interface {
	Validate() error
}
