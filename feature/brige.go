package feature

import (
	"breaker/pkg/protocol"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"strconv"
)

const FBridge = "bridge"

type BridgeConfig struct {
	ServerAddr string `ini:"server_addr"`
	LocalPort  int    `ini:"local_port"`
	RemotePort int    `ini:"remote_port"`
	ProxyName  string `ini:"proxy_name"`
}

func init() {
	RegisterConfig(FBridge, &BridgeConfig{})
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
	portal     net.Conn
}

func (b *Bridge) Stop(ctx context.Context) error {
	if b.portal != nil {
		err := protocol.WriteMsg(b.portal, &protocol.CloseProxy{
			ProxyName: b.ProxyName,
		})
		if err != nil {
			log.Errorf("send CloseProxy error:[%s]", err.Error())
		}
	}
	err := b.portal.Close()
	b.portal = nil
	return err
}

func (b *Bridge) Name() string {
	return FBridge
}

func (b *Bridge) Connect() error {
	if b.portal != nil {
		return errors.New("already connect to portal:" + b.portal.RemoteAddr().String())
	}
	log.Info("dial tcp:", b.ServerAddr)
	portal, err := net.Dial("tcp", b.ServerAddr)

	if err != nil {
		return err
	}
	b.portal = portal
	newProxy := &protocol.NewProxy{
		RemotePort: b.RemotePort,
		ProxyName:  b.ProxyName,
	}
	log.Info("log message:", newProxy)
	err = protocol.WriteMsg(portal, newProxy)
	if err != nil {
		log.Errorf("send new proxy error:[%s]", err.Error())
		return err
	}
	_, err = protocol.ReadResponse(portal)
	if err != nil {
		log.Errorf("ReadResponse:[%s]", err.Error())
		return err
	}

	//read newProxyResponse
	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(b.LocalPort))
	local, err := net.Dial("tcp", addr)
	log.Info("dial tcp:", addr)
	if err != nil {
		_ = protocol.WriteMsg(portal, &protocol.CloseProxy{
			ProxyName: b.ProxyName,
		})
		_, _ = protocol.ReadResponse(portal)
		return err
	}
	workCtl := &protocol.WorkCtl{
		ProxyName: b.ProxyName,
	}
	err = protocol.WriteMsg(portal, workCtl)
	if err != nil {
		return err
	}
	_, err = protocol.ReadResponse(portal)
	if err != nil {
		return err
	}

	go io.Copy(local, portal)
	io.Copy(portal, local)
	return nil
}
