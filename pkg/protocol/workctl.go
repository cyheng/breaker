package protocol

type WorkCtl struct {
	TraceID string
}

func (n *WorkCtl) Type() byte {
	return TypeNewWorkCtl
}

func init() {
	RegisterCommand(&WorkCtl{})
}
