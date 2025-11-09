package clustertestutils

import (
	"errors"
)

// Common error variables used across cluster provisioner tests to avoid duplication.
var (
	ErrCreateClusterFailed  = errors.New("create cluster failed")
	ErrDeleteClusterFailed  = errors.New("delete cluster failed")
	ErrListClustersFailed   = errors.New("list clusters failed")
	ErrStartClusterFailed   = errors.New("start cluster failed")
	ErrStopClusterFailed    = errors.New("stop cluster failed")
	ErrScaleNodeGroupFailed = errors.New("scale node group failed")
)
