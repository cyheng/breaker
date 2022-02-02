package protocol

type NewMaster struct {
}

func (n *NewMaster) Type() byte {
	return TypeNewMaster
}

type NewMasterResp struct {
	Resp
	TraceID string
}

func (n *NewMasterResp) Type() byte {
	return TypeNewMasterResp
}
func init() {
	RegisterCommand(&NewMaster{})
	RegisterCommand(&NewMasterResp{})
}
