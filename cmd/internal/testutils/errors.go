// Package testutils provides testing utilities and shared test errors for command tests.
package testutils

import "errors"

// ErrTestConfigLoadError is a static test error to comply with err113.
var ErrTestConfigLoadError = errors.New("test config load error")
