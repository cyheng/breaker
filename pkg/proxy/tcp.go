package proxy

import (
	"net"
)

type TcpProxy struct {
	Name string
	net.Listener
	CreateBy net.Conn
}

func NewTcpProxy(name string, listener net.Listener, CreateBy net.Conn) *TcpProxy {
	return &TcpProxy{
		Name:     name,
		Listener: listener,
		CreateBy: CreateBy,
	}
}
