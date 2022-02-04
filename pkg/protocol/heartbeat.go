package protocol

type Ping struct {
}

func (n *Ping) Type() byte {
	return TypePing
}

type Pong struct {
	Resp
}

func (n *Pong) Type() byte {
	return TypePong
}
