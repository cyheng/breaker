package feature

import (
	"breaker/pkg/protocol"
	"breaker/pkg/protocol/command"
	"context"
	"io"
	"net"
	"strconv"
)

type BridgeConfig struct {
	ServerAddr string `ini:"server_addr"`
	LocalPort  int    `ini:"local_port"`
	RemotePort int    `ini:"remote_port"`
	ProxyName  string `ini:"proxy_name"`
}

func (b *BridgeConfig) OnInit() {
	if b.ServerAddr == "" {
		panic("server address can not be empty")
	}
	if b.LocalPort < 0 || b.LocalPort > 65535 {
		panic("invalid local port[0-65535]")
	}
	if b.RemotePort < 0 || b.RemotePort > 65535 {
		panic("invalid local port[0-65535]")
	}
	if b.ProxyName == "" {
		b.ProxyName = b.ServerAddr + "_to_" + strconv.Itoa(b.LocalPort)
	}
}

func (b *BridgeConfig) NewFeature() (Feature, error) {
	bridge := &Bridge{}
	bridge.ServerAddr = b.ServerAddr
	bridge.LocalPort = b.LocalPort
	bridge.RemotePort = b.RemotePort
	bridge.ProxyName = b.ProxyName
	return bridge, nil
}

type Bridge struct {
	ServerAddr string
	LocalPort  int
	RemotePort int
	ProxyName  string
}

func (b *Bridge) Stop(ctx context.Context) error {
	return nil
}

func (b *Bridge) Name() string {
	return "bridge"
}

func (b *Bridge) Connect() error {
	portal, err := net.Dial("tcp", b.ServerAddr)
	if err != nil {
		return err
	}
	newProxy := &command.NewProxy{
		RemotePort: b.RemotePort,
		ProxyName:  b.ProxyName,
	}
	err = protocol.WriteMsg(portal, newProxy)
	if err != nil {
		return err
	}
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(b.LocalPort))
	local, err := net.Dial("tcp", addr)
	if err != nil {
		_ = protocol.WriteMsg(portal, &command.CloseProxy{
			ProxyName: b.ProxyName,
		})
		return err
	}
	workCtl := &command.WorkCtl{
		ProxyName: b.ProxyName,
	}
	err = protocol.WriteMsg(portal, workCtl)
	if err != nil {
		return err
	}
	go io.Copy(local, portal)
	io.Copy(portal, local)
	return nil
}
