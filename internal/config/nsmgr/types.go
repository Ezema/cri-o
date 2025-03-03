//go:build linux
// +build linux

package nsmgr

import (
	"os"
	"sync"

	nspkg "github.com/containernetworking/plugins/pkg/ns"
	"github.com/containers/storage/pkg/idtools"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// NSType is a representation of available namespace types.
type NSType string

const (
	NETNS                NSType = "net"
	IPCNS                NSType = "ipc"
	UTSNS                NSType = "uts"
	USERNS               NSType = "user"
	PIDNS                NSType = "pid"
	ManagedNamespacesNum        = 4
)

// supportedNamespacesForPinning returns a slice of
// the names of namespaces that CRI-O supports
// pinning.
func supportedNamespacesForPinning() []NSType {
	return []NSType{NETNS, IPCNS, UTSNS, USERNS}
}

type PodNamespacesConfig struct {
	Namespaces []*PodNamespaceConfig
	IDMappings *idtools.IDMappings
	Sysctls    map[string]string
}

type PodNamespaceConfig struct {
	Type NSType
	Host bool
	Path string
}

// Namespace provides a generic namespace interface.
type Namespace interface {
	// Path returns the bind mount path of the namespace.
	Path() string

	// Type returns the namespace type (net, ipc, user, pid or uts).
	Type() NSType

	// Remove ensures this namespace is closed and removed.
	Remove() error
}

// namespace is the internal implementation of the Namespace interface.
type namespace struct {
	sync.Mutex
	ns     NS
	closed bool
	nsType NSType
	nsPath string
}

// NS is a wrapper for the containernetworking plugin's NetNS interface
// It exists because while NetNS is specifically called such, it is really a generic
// namespace, and can be used for other namespace types.
type NS interface {
	nspkg.NetNS
}

// Path returns the bind mount path of the namespace.
func (n *namespace) Path() string {
	if n == nil || n.ns == nil {
		return ""
	}
	return n.nsPath
}

// Type returns the namespace type (net, ipc, user, pid or uts).
func (n *namespace) Type() NSType {
	return n.nsType
}

// Remove ensures this namespace is closed and removed.
func (n *namespace) Remove() error {
	n.Lock()
	defer n.Unlock()

	if n.closed {
		// Remove() can be called multiple
		// times without returning an error.
		return nil
	}

	if err := n.ns.Close(); err != nil {
		return err
	}

	n.closed = true

	fp := n.Path()
	if fp == "" {
		return nil
	}

	// try to unmount, ignoring "not mounted" (EINVAL) error.
	if err := unix.Unmount(fp, unix.MNT_DETACH); err != nil && err != unix.EINVAL {
		return errors.Wrapf(err, "unable to unmount %s", fp)
	}
	return os.RemoveAll(fp)
}

// GetNamespace takes a path and type, checks if it is a namespace, and if so
// returns an instance of the Namespace interface.
func GetNamespace(nsPath string, nsType NSType) (Namespace, error) {
	ns, err := nspkg.GetNS(nsPath)
	if err != nil {
		return nil, err
	}

	return &namespace{ns: ns, nsType: nsType, nsPath: nsPath}, nil
}
