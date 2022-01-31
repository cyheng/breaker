package server

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

var ErrServerStopped = errors.New("server stopped")

const (
	tempErrDelay             = time.Millisecond * 5
	QueueSize                = 1024
	DefaultWriteAttemptTimes = 1
)

type Server struct {
	listener net.Listener
	// Packer is the message packer, will be passed to session.
	Packer Packer

	// Codec is the message codec, will be passed to session.
	Codec Codec

	// OnSessionCreate is an event hook, will be invoked when session's created.
	OnSessionCreate func(sess Session)

	// OnSessionClose is an event hook, will be invoked when session's closed.
	OnSessionClose func(sess Session)

	socketReadBufferSize  int
	socketWriteBufferSize int
	readTimeout           time.Duration
	writeTimeout          time.Duration
	respQueueSize         int
	router                *Router
	stopped               chan struct{}
	writeAttemptTimes     int
}

func NewServer(opts ...Option) *Server {
	srv := &Server{

		stopped:           make(chan struct{}),
		Packer:            NewDefaultPacker(),
		Codec:             NewDefaultCodec(),
		router:            NewRouter(),
		respQueueSize:     QueueSize,
		writeAttemptTimes: DefaultWriteAttemptTimes,
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}
func (s *Server) Serve(addr string) error {
	address, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}
	lis, err := net.ListenTCP("tcp", address)
	if err != nil {
		return err
	}
	s.listener = lis
	log.Infof("start tcp server:" + addr)
	return s.acceptLoop()
}

func (s *Server) acceptLoop() error {
	for {
		if s.IsStopped() {
			log.Tracef("server accept loop stopped")
			return ErrServerStopped
		}

		conn, err := s.listener.Accept()
		if err != nil {
			if s.IsStopped() {
				log.Tracef("server accept loop stopped")
				return ErrServerStopped
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Errorf("accept err: %s; retrying in %s", err, tempErrDelay)
				time.Sleep(tempErrDelay)
				continue
			}
			return fmt.Errorf("accept err: %s", err)
		}
		if s.socketReadBufferSize > 0 {
			if err := conn.(*net.TCPConn).SetReadBuffer(s.socketReadBufferSize); err != nil {
				return fmt.Errorf("conn set read buffer err: %s", err)
			}
		}
		if s.socketWriteBufferSize > 0 {
			if err := conn.(*net.TCPConn).SetWriteBuffer(s.socketWriteBufferSize); err != nil {
				return fmt.Errorf("conn set write buffer err: %s", err)
			}
		}

		go s.handleConn(conn)
	}
}

func (s *Server) IsStopped() bool {
	select {
	case <-s.stopped:
		return true
	default:
		return false
	}
}

func (s *Server) handleConn(conn net.Conn) {
	session := NewTcpSession(conn,
		AsCodec(s.Codec),
		AsPacker(s.Packer),
		AsQueueSize(s.respQueueSize),
	)
	if s.OnSessionCreate != nil {
		go s.OnSessionCreate(session)
	}

	go session.readInbound(s.router, s.readTimeout)               // start reading message packet from connection.
	go session.writeOutbound(s.writeTimeout, s.writeAttemptTimes) // start writing message packet to connection.

	select {
	case <-session.closed: // wait for session finished.
	case <-s.stopped: // or the server is stopped.
	}

	if s.OnSessionClose != nil {
		go s.OnSessionClose(session)
	}
}
func (s *Server) AddRoute(msgID byte, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	s.router.register(msgID, handler, middlewares...)
}

// Use registers global middlewares to the router.
func (s *Server) Use(middlewares ...MiddlewareFunc) {
	s.router.registerMiddleware(middlewares...)
}

// NotFoundHandler sets the not-found handler for router.
func (s *Server) NotFoundHandler(handler HandlerFunc) {
	s.router.setNotFoundHandler(handler)
}

// Stop stops server. Closing Listener and all connections.
func (s *Server) Stop() error {
	close(s.stopped)
	return s.listener.Close()
}

type Option func(*Server)

func WithPacker(pack Packer) Option {
	return func(server *Server) {
		server.Packer = pack
	}
}

func WithCodec(codec Codec) Option {
	return func(server *Server) {
		server.Codec = codec
	}
}

func WithOnSessionCreate(fn func(sess Session)) Option {
	return func(server *Server) {
		server.OnSessionCreate = fn
	}
}
func WithOnSessionClose(fn func(sess Session)) Option {
	return func(server *Server) {
		server.OnSessionClose = fn
	}
}
func WithSocketReadBufferSize(readSize int) Option {
	return func(server *Server) {
		server.socketReadBufferSize = readSize
	}
}
func WithSocketWriteBufferSize(writeSize int) Option {
	return func(server *Server) {
		server.socketWriteBufferSize = writeSize
	}
}
func WithReadTimeOut(readTimeout time.Duration) Option {
	return func(s *Server) {
		s.readTimeout = readTimeout
	}

}

func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(s *Server) {
		s.writeTimeout = writeTimeout

	}
}

func WithWriteAttemptTimes(writeAttemptTimes int) Option {
	return func(s *Server) {
		s.writeAttemptTimes = writeAttemptTimes
	}
}
func WithRespQueueSize(respQueueSize int) Option {
	return func(s *Server) {
		s.respQueueSize = respQueueSize
	}
}
