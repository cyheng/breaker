package protocol

type CloseProxy struct {
	ProxyName string
}

func (n *CloseProxy) Type() byte {
	return TypeCloseProxy
}

func init() {
	RegisterCommand(&CloseProxy{})
}
