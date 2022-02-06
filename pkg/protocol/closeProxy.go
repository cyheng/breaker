package protocol

type CloseProxy struct {
	ProxyName string
}

func (n *CloseProxy) Type() byte {
	return TypeCloseProxy
}

type CloseProxyResp struct {
	Resp
	ProxyName string
}

func (n *CloseProxyResp) Type() byte {
	return TypeCloseProxyResp
}
