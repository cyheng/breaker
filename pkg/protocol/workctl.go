package protocol

const TypeNewWorkCtl = 'w'

type WorkCtl struct {
	ProxyName string
}

func (n *WorkCtl) Type() byte {
	return TypeNewWorkCtl
}

func init() {
	RegisterCommand(&WorkCtl{})
}
