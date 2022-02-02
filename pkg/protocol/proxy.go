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
}

func (n *NewProxyResp) Type() byte {
	return TypeNewProxyResp
}
func init() {
	RegisterCommand(&NewProxy{})
	RegisterCommand(&NewProxyResp{})
}
