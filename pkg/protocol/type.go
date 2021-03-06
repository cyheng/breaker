package protocol

const (
	TypeCloseProxy     = '1'
	TypeNewProxy       = '2'
	TypeNewWorkCtl     = '3'
	TypeNewMaster      = '4'
	TypeReqWorkCtl     = '5'
	TypeCloseProxyResp = '6'
	TypeNewProxyResp   = '7'
	TypeNewWorkCtlResp = '8'
	TypeNewMasterResp  = '9'
	TypeReqWorkCtlResp = '0'
	TypePing           = 'a'
	TypePong           = 'b'
)

func init() {
	RegisterCommand(&CloseProxy{})
	RegisterCommand(&CloseProxyResp{})
	RegisterCommand(&NewMaster{})
	RegisterCommand(&NewMasterResp{})
	RegisterCommand(&NewProxy{})
	RegisterCommand(&NewProxyResp{})
	RegisterCommand(&NewWorkCtl{})
	RegisterCommand(&ReqWorkCtl{})
	RegisterCommand(&ReqWorkCtlResp{})
	RegisterCommand(&NewWorkCtlResp{})
	RegisterCommand(&Ping{})
	RegisterCommand(&Pong{})
}
