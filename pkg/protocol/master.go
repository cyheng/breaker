package protocol

type NewMaster struct {
}

func (n *NewMaster) Type() byte {
	return TypeNewMaster
}

type NewMasterResp struct {
	Resp
	SessionId string
}

func (n *NewMasterResp) Type() byte {
	return TypeNewMasterResp
}
