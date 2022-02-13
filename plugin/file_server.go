package plugin

import (
	"breaker/pkg/errwrap"
	"errors"
	"net"
	"net/http"
	"sync"
)

type FileServer struct {
	FileLocation string
	Prefix       string
	listener     *Listener
	server       *http.Server
}

func NewFileServer(fileLocation string, prefix string) *FileServer {
	return &FileServer{FileLocation: fileLocation, Prefix: prefix}
}

func (s *FileServer) Run() error {
	router := http.NewServeMux()
	handler := http.StripPrefix(s.Prefix, http.FileServer(http.Dir(s.FileLocation)))

	router.Handle(s.Prefix, handler)
	server := &http.Server{
		Handler: router,
	}
	s.listener = newListener()
	s.server = server
	return server.Serve(s.listener)
}

func (s *FileServer) HandlerConn(remote, local net.Conn) error {
	return s.listener.AddConn(remote)
}

func (s *FileServer) Close() error {
	err := errwrap.PanicToError(func() {
		s.server.Close()
		s.listener.Close()
	})
	return err
}

type Listener struct {
	conn chan net.Conn
	once sync.Once
}

func newListener() *Listener {
	return &Listener{
		conn: make(chan net.Conn, 128),
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	conn, ok := <-l.conn
	if !ok {
		return nil, errors.New("listener closed")
	}
	return conn, nil
}
func (l *Listener) AddConn(c net.Conn) error {
	err := errwrap.PanicToError(func() {
		l.conn <- c
	})
	return err
}
func (l *Listener) Close() error {
	l.once.Do(func() {
		close(l.conn)
	})
	return nil
}

func (l *Listener) Addr() net.Addr {
	return nil
}
