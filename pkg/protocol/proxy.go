package protocol

type NewProxy struct {
	RemotePort int
	ProxyName  string
	TraceId    string
}

func (n *NewProxy) Type() byte {
	return TypeNewProxy
}

type NewProxyResp struct {
	Resp
	ProxyName string
}

func (n *NewProxyResp) Type() byte {
	return TypeNewProxyResp
}
