package device

import (
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

type Device interface {
	stack.LinkEndpoint
	Close() error
}
