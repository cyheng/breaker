package server

import (
	"breaker/pkg/uuid"
	"net"
	"sync"
	"time"
)

type Session interface {
	// ID returns current session's id.
	ID() interface{}

	// SetID sets current session's id.
	SetID(id interface{})

	// Send sends the ctx to the respQueue.
	Send(ctx Context) bool

	// Codec returns the codec, can be nil.
	Codec() Codec

	// Close closes current session.
	Close()

	// AllocateContext gets a Context ships with current session.
	AllocateContext() Context
}

type TcpSession struct {
	id        string        // session's ID.
	conn      net.Conn      // tcp connection
	closed    chan struct{} // to close()
	closeOnce sync.Once     // ensure one session only close once
	respQueue chan Context  // response queue channel, pushed in Send() and popped in writeOutbound()
	packer    Packer        // to pack and unpack message
	codec     Codec         // encode/decode message data
	ctxPool   sync.Pool     // router context pool
}

type SessionOpt func(*TcpSession)

func NewTcpSession(conn net.Conn, ops ...SessionOpt) *TcpSession {
	var session = &TcpSession{
		id:        uuid.New().String(),
		conn:      conn,
		closed:    make(chan struct{}),
		respQueue: make(chan Context, QueueSize),
		ctxPool:   sync.Pool{New: func() interface{} { return NewContext() }},
	}
	for _, op := range ops {
		op(session)
	}
	return session
}

func (s *TcpSession) ID() interface{} {
	//TODO implement me
	panic("implement me")
}

func (s *TcpSession) SetID(id interface{}) {
	//TODO implement me
	panic("implement me")
}

func (s *TcpSession) Send(ctx Context) bool {
	//TODO implement me
	panic("implement me")
}

func (s *TcpSession) Codec() Codec {
	//TODO implement me
	panic("implement me")
}

func (s *TcpSession) Close() {
	//TODO implement me
	panic("implement me")
}

func (s *TcpSession) AllocateContext() Context {
	//TODO implement me
	panic("implement me")
}

func (s *TcpSession) readInbound(router *Router, timeout time.Duration) {

}

func (s *TcpSession) writeOutbound(timeout time.Duration, times int) {

}

func AsPacker(packer Packer) SessionOpt {
	return func(s *TcpSession) {
		s.packer = packer
	}

}

func AsCodec(codec Codec) SessionOpt {
	return func(s *TcpSession) {
		s.codec = codec
	}
}
func AsQueueSize(queueSize int) SessionOpt {
	return func(s *TcpSession) {
		s.respQueue = make(chan Context, queueSize)
	}
}
