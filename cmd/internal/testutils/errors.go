package testutils

import "errors"

// ErrTestConfigLoadError is a static test error to comply with err113.
var ErrTestConfigLoadError = errors.New("test config load error")
