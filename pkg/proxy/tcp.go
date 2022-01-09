package proxy

import (
	"net"
)

type TcpProxy struct {
	Name string
	net.Listener
	WorkingChan chan net.Conn
}

func NewTcpProxy(name string, listener net.Listener) *TcpProxy {
	return &TcpProxy{
		Name:        name,
		Listener:    listener,
		WorkingChan: make(chan net.Conn, 10),
	}
}

func (t *TcpProxy) Close() {
	t.Listener.Close()
	close(t.WorkingChan)
}
