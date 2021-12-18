package command

import "breaker/pkg/protocol"

const TypeNewProxy = 'p'

type NewProxy struct {
	RemotePort int
	ProxyName  string
}

func (n *NewProxy) Type() byte {
	return TypeNewProxy
}

func init() {
	protocol.RegisterCommand(&NewProxy{})
}
