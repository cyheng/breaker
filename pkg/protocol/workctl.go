package protocol

type WorkCtl struct {
	TraceID   string
	ProxyName string
}

func (n *WorkCtl) Type() byte {
	return TypeNewWorkCtl
}

type ReqWorkCtl struct {
	ProxyName string
}

func (n *ReqWorkCtl) Type() byte {
	return TypeReqWorkCtl
}
func init() {
	RegisterCommand(&WorkCtl{})
	RegisterCommand(&ReqWorkCtl{})
}
