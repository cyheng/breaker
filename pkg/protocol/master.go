package protocol

type NewMaster struct {
	RemotePort int
	ProxyName  string
}

func (n *NewMaster) Type() byte {
	return TypeNewMaster
}

func init() {
	RegisterCommand(&NewMaster{})
}
