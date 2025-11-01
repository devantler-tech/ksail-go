package gen //nolint:testpackage // Tests need access to unexported helpers

// NOTE: Ingress command returns a parent command with no direct test.
// Ingress resources would typically be created via kubectl with specific
// rule formats that may vary by Kubernetes version and environment.
// Users should test ingress generation manually as needed.
