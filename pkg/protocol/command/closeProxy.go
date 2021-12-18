package command

import "breaker/pkg/protocol"

const TypeCloseProxy = 'c'

type CloseProxy struct {
	ProxyName string
}

func (n *CloseProxy) Type() byte {
	return TypeCloseProxy
}

func init() {
	protocol.RegisterCommand(&CloseProxy{})
}
