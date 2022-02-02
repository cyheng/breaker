package breaker

import (
	"breaker/pkg/protocol"
	"breaker/pkg/uuid"
	log "github.com/sirupsen/logrus"
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

	Conn() net.Conn

	SendSync(context Context) bool
}

type TcpSession struct {
	id        interface{}   // session's ID.
	conn      net.Conn      // tcp connection
	closed    chan struct{} // to close()
	closeOnce sync.Once     // ensure one session only close once
	respQueue chan Context  // response queue channel, pushed in Send() and popped in writeOutbound()
	packer    Packer        // to pack and unpack message
	codec     Codec         // encode/decode message data
	ctxPool   sync.Pool     // router context pool
}

func (s *TcpSession) Conn() net.Conn {
	return s.conn
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
	return s.id
}

func (s *TcpSession) SetID(id interface{}) {
	s.id = id
}
func (s *TcpSession) SendCmd(cmd protocol.Command) bool {
	ctx := s.AllocateContext()
	return ctx.SetResponseMessage(cmd).Send()
}
func (s *TcpSession) Send(ctx Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-s.closed:
		return false
	case s.respQueue <- ctx:
		return true
	}
}

func (s *TcpSession) SendSync(ctx Context) bool {
	select {
	case <-ctx.Done():
		return false
	case <-s.closed:
		return false
	default:
	}
	outboundMsg, err := s.packResponse(ctx)
	if err != nil {
		log.Errorf("session %s pack outbound message err: %s", s.id, err)
		return false
	}
	if outboundMsg == nil {
		return false
	}
	if err := s.attemptConnWrite(outboundMsg, 1); err != nil {
		log.Errorf("session %s conn write err: %s", s.id, err)
		return false
	}
	return true
}

func (s *TcpSession) Codec() Codec {
	return s.codec
}

func (s *TcpSession) Close() {
	s.closeOnce.Do(func() { close(s.closed) })
}

func (s *TcpSession) AllocateContext() Context {
	c := s.ctxPool.Get().(*routeContext)
	c.reset()
	c.SetSession(s)
	return c
}

func (s *TcpSession) readInbound(router *Router, timeout time.Duration) {
	for {
		select {
		case <-s.closed:
			return
		default:
		}
		if timeout > 0 {
			if err := s.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				log.Errorf("session %s set read deadline err: %s", s.id, err)
				break
			}
		}
		reqEntry, err := s.packer.Unpack(s.conn)
		if err != nil {
			log.Errorf("session %s unpack inbound packet err: %s", s.id, err)
			break
		}
		if reqEntry == nil {
			continue
		}

		s.handleReq(router, reqEntry)
	}
	log.Tracef("session %s readInbound exit because of error", s.id)
	s.Close()
}

func (s *TcpSession) writeOutbound(timeout time.Duration, times int) {
	for {
		var ctx Context
		select {
		case <-s.closed:
			return

		case ctx = <-s.respQueue:
		}

		outboundMsg, err := s.packResponse(ctx)
		if err != nil {
			log.Errorf("session %s pack outbound message err: %s", s.id, err)
			continue
		}
		if outboundMsg == nil {
			continue
		}

		if timeout > 0 {
			if err := s.conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
				log.Errorf("session %s set write deadline err: %s", s.id, err)
				break
			}
		}

		if err := s.attemptConnWrite(outboundMsg, times); err != nil {
			log.Errorf("session %s conn write err: %s", s.id, err)
			break
		}
	}
	s.Close()
	log.Tracef("session %s writeOutbound exit because of error", s.id)
}

func (s *TcpSession) handleReq(router *Router, entry protocol.Command) {
	ctx := s.AllocateContext().SetRequestMessage(entry)
	router.handleRequest(ctx)
	s.Send(ctx)
}

func (s *TcpSession) attemptConnWrite(outboundMsg []byte, attemptTimes int) (err error) {
	for i := 0; i < attemptTimes; i++ {
		time.Sleep(tempErrDelay * time.Duration(i))
		_, err = s.conn.Write(outboundMsg)

		// breaks if err is not nil or it's the last attempt.
		if err == nil || i == attemptTimes-1 {
			break
		}

		// check if err is `net.Error`
		ne, ok := err.(net.Error)
		if !ok {
			break
		}
		if ne.Timeout() {
			break
		}
		if ne.Temporary() {
			log.Errorf("session %s conn write err: %s; retrying in %s", s.id, err, tempErrDelay*time.Duration(i+1))
			continue
		}
		break // if err is not temporary, break the loop.
	}
	return
}

func (s *TcpSession) packResponse(ctx Context) ([]byte, error) {
	defer s.ctxPool.Put(ctx)
	if ctx.Response() == nil {
		return nil, nil
	}
	return s.packer.Pack(ctx.Response())
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
