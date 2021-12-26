package protocol

type NewProxy struct {
	RemotePort int
	ProxyName  string
}

func (n *NewProxy) Type() byte {
	return TypeNewProxy
}

func init() {
	RegisterCommand(&NewProxy{})
}
