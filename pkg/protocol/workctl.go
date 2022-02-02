package protocol

type NewWorkCtl struct {
	TraceID   string
	ProxyName string
}

func (n *NewWorkCtl) Type() byte {
	return TypeNewWorkCtl
}

type NewWorkCtlResp struct {
	Resp
	TraceID   string
	ProxyName string
}

func (n *NewWorkCtlResp) Type() byte {
	return TypeNewWorkCtlResp
}

type ReqWorkCtl struct {
	ProxyName string
}

func (n *ReqWorkCtl) Type() byte {
	return TypeReqWorkCtl
}

type ReqWorkCtlResp struct {
	Resp
	ProxyName string
}

func (n *ReqWorkCtlResp) Type() byte {
	return TypeReqWorkCtlResp
}
func init() {
	RegisterCommand(&NewWorkCtl{})
	RegisterCommand(&ReqWorkCtl{})
	RegisterCommand(&ReqWorkCtlResp{})
	RegisterCommand(&NewWorkCtlResp{})
}
