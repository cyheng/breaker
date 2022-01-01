package protocol

type WorkCtl struct {
	TraceID string
}

func (n *WorkCtl) Type() byte {
	return TypeNewWorkCtl
}

type ReqWorkCtl struct {
}

func (n *ReqWorkCtl) Type() byte {
	return TypeReqWorkCtl
}
func init() {
	RegisterCommand(&WorkCtl{})
	RegisterCommand(&ReqWorkCtl{})
}
