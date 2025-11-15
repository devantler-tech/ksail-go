package flannel

import _ "embed"

// KindBaseCNIPluginsManifest is the DaemonSet manifest that installs
// the base CNI plugin binaries (bridge, host-local, loopback, portmap, ptp, etc.)
// onto Kind nodes before Flannel installation when disableDefaultCNI is true.
//
//go:embed cni-bootstrap.yaml
var KindBaseCNIPluginsManifest string
