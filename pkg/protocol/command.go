package protocol

type Command interface {
	Type() byte
}

type Resp struct {
	Error string
}
