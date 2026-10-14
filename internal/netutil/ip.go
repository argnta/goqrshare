package netutil

import "net"

type IPCandidate struct {
	Name      string
	IP        net.IP
	IsPrivate bool
	IsVirtual bool
}
