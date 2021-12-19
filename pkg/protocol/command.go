package protocol

type Command interface {
	Type() byte
}
