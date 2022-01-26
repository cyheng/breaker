package server

import (
	"breaker/pkg/protocol"
	"io"
)

type Packer interface {
	// Pack packs Message into the packet to be written.
	Pack(cmd protocol.Command) ([]byte, error)

	// Unpack unpacks the message packet from reader,
	// returns the protocol.Command, and error if error occurred.
	Unpack(reader io.Reader) (protocol.Command, error)
}

type DefaultPacker struct{}

func NewDefaultPacker() *DefaultPacker {
	return &DefaultPacker{}
}

func (p *DefaultPacker) Pack(cmd protocol.Command) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}

func (p *DefaultPacker) Unpack(reader io.Reader) (protocol.Command, error) {
	//TODO implement me
	panic("implement me")
}
