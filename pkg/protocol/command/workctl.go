package command

import "breaker/pkg/protocol"

const TypeNewWorkCtl = 'w'

type WorkCtl struct {
	ProxyName string
}

func (n *WorkCtl) Type() byte {
	return TypeNewProxy
}

func init() {
	protocol.RegisterCommand(&WorkCtl{})
}
