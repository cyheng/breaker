package services

import (
	"breaker/feature"
	"breaker/pkg/protocol"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"runtime/debug"
	"strconv"
)

type Bridge struct {
	ServerAddr  string
	LocalPort   int
	RemotePort  int
	ProxyName   string
	portal      net.Conn
	traceId     string
	msgReadChan chan protocol.Command
}

func NewBridge() Service {
	bridge := &Bridge{
		msgReadChan: make(chan protocol.Command, 10),
	}

	return bridge
}
func (b *Bridge) Init(cfg *feature.BridgeConfig) {
	b.ServerAddr = cfg.ServerAddr
	b.LocalPort = cfg.LocalPort
	b.RemotePort = cfg.RemotePort
	b.ProxyName = cfg.ProxyName
}
func (b *Bridge) Stop(ctx context.Context) {
	if b.portal != nil {
		err := protocol.WriteMsg(b.portal, &protocol.CloseProxy{
			ProxyName: b.ProxyName,
		})
		if err != nil {
			log.Errorf("send CloseProxy error:[%s]", err.Error())
		}
		defer b.portal.Close()
	}
	close(b.msgReadChan)
	b.portal = nil

}

func (b *Bridge) Start(args interface{}, ctx context.Context) error {
	cfg := args.(*feature.BridgeConfig)
	b.Init(cfg)
	if b.portal != nil {
		return errors.New("already connect to portal:" + b.portal.RemoteAddr().String())
	}

	portal, traceId, err := b.newMaster()
	if err != nil {
		return err
	}

	//read send new proxy
	newProxy := &protocol.NewProxy{

		RemotePort: b.RemotePort,
		ProxyName:  b.ProxyName,
	}
	log.Infof("send message:[NewProxy]")
	_, err = sendMsgAndWaitResponse(portal, newProxy)
	if err != nil {
		return err
	}
	egg, _ := errgroup.WithContext(ctx)
	egg.Go(func() error {
		return b.SendWorkerConn(traceId, ctx)
	})
	egg.Go(func() error {
		return b.getMessageFromPortal(portal, ctx)
	})

	egg.Go(func() error {
		return b.msgHandler(ctx)
	})

	return egg.Wait()
}

func (b *Bridge) newMaster() (net.Conn, string, error) {
	log.Info("dial tcp:", b.ServerAddr)
	portal, err := net.Dial("tcp", b.ServerAddr)
	if err != nil {
		return nil, "", err
	}
	b.portal = portal

	newMaster := &protocol.NewMaster{}
	log.Infof("send message:[master]")
	res, err := sendMsgAndWaitResponse(portal, newMaster)
	if err != nil {
		return nil, "", err
	}
	traceId := res.Data.(string)
	b.traceId = traceId
	return portal, traceId, err
}

func (b *Bridge) SendWorkerConn(traceId string, ctx context.Context) error {

	workCtl := &protocol.WorkCtl{
		TraceID: traceId,
	}
	log.Infof("send message:[workCtl]")

	workerConn, err := net.Dial("tcp", b.ServerAddr)
	if err != nil {
		return err
	}
	defer workerConn.Close()
	_, err = sendMsgAndWaitResponse(workerConn, workCtl)
	if err != nil {
		return err
	}

	addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(b.LocalPort))
	log.Info("dial tcp:", addr)
	local, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer local.Close()
	go io.Copy(local, workerConn)

	io.Copy(workerConn, local)
	return nil
}

func (b *Bridge) getMessageFromPortal(portal net.Conn, ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()
	for {
		cmd, err := protocol.ReadMsg(portal)
		if err != nil {
			if err == io.EOF {
				log.Debug("control connection closed")
				return err
			}
			log.Warn("read error: %v", err)

			return err
		}

		b.msgReadChan <- cmd
	}

}

func (b *Bridge) msgHandler(ctx context.Context) (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()
	for {
		select {
		case msg := <-b.msgReadChan:
			switch msg.(type) {
			case *protocol.ReqWorkCtl:
				go b.SendWorkerConn(b.traceId, ctx)
				break
			default:
				log.Debug("unknown command")
				err = errors.New("unknown command")
			}
		}

	}

}

func sendMsgAndWaitResponse(portal net.Conn, cmd protocol.Command) (res *protocol.Response, err error) {

	err = protocol.WriteMsg(portal, cmd)
	if err != nil {
		return nil, err
	}
	res, err = protocol.ReadResponse(portal)
	if err != nil {
		log.Errorf("ReadResponse:[%s]", err.Error())
		return nil, err
	}
	return res, err
}
