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
	ServerAddr string
	LocalPort  int
	RemotePort int
	ProxyName  string
	master     net.Conn
	traceId    string
	//写入的goroutine中关闭chan
	msgReadChan  chan protocol.Command
	msgWriteChan chan protocol.Command
	workingChan  chan string
}

func NewBridge() Service {
	bridge := &Bridge{
		msgReadChan:  make(chan protocol.Command, 10),
		msgWriteChan: make(chan protocol.Command, 10),
		workingChan:  make(chan string, 10),
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
	if b.master != nil {
		err := protocol.WriteMsg(b.master, &protocol.CloseProxy{
			ProxyName: b.ProxyName,
		})
		if err != nil {
			log.Errorf("send CloseProxy error:[%s]", err.Error())
		}
		defer b.master.Close()
	}
	close(b.msgReadChan)
	b.master = nil

}

func (b *Bridge) Start(args interface{}, ctx context.Context) error {
	cfg := args.(*feature.BridgeConfig)
	b.Init(cfg)
	if b.master != nil {
		return errors.New("already connect to master:" + b.master.RemoteAddr().String())
	}
	portal, traceId, err := getMaster(b.ServerAddr)
	if err != nil {
		return err
	}
	b.master = portal
	b.traceId = traceId

	//send new proxy
	newProxy := &protocol.NewProxy{
		RemotePort: b.RemotePort,
		ProxyName:  b.ProxyName,
	}
	b.msgWriteChan <- newProxy
	log.Infof("send new proxy:%s :%d", b.ProxyName, b.RemotePort)
	ctx, cancel := context.WithCancel(ctx)
	egg, _ := errgroup.WithContext(ctx)
	egg.Go(func() error {
		return b.SendWorkerConn(ctx)
	})
	egg.Go(func() error {
		return b.readHandler(portal, ctx)
	})
	egg.Go(func() error {
		return b.writeHandler(portal, ctx)
	})
	egg.Go(func() error {
		return b.msgHandler(ctx)
	})
	for i := 0; i < 3; i++ {
		b.workingChan <- b.ProxyName
	}
	err = egg.Wait()
	if err != nil {
		cancel()
		return err
	}
	return nil
}

func getMaster(addr string) (net.Conn, string, error) {
	//get master connection
	log.Info("dial tcp:", addr)
	portal, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, "", err
	}
	newMaster := &protocol.NewMaster{}
	log.Infof("send message:[master]")
	res, err := sendMsgAndWaitResponse(portal, newMaster)
	if err != nil {
		return nil, "", err
	}
	traceId := res.Data.(string)
	log.Infof("get master with traceid:[%s]", traceId)
	return portal, traceId, nil
}

func (b *Bridge) SendWorkerConn(ctx context.Context) error {
	for {
		select {
		case pxy := <-b.workingChan:
			go b.createWorker(ctx, pxy)
		case <-ctx.Done():
			return ctx.Err()

		}
	}
}

func (b *Bridge) createWorker(ctx context.Context, pxy string) (err error) {
	defer func() {
		if err != nil {
			log.Errorf("error:[%+v]", err)
		}
	}()
	if b.ProxyName != pxy {
		return errors.New("pxy not found")
	}
	workCtl := &protocol.WorkCtl{
		TraceID:   b.traceId,
		ProxyName: pxy,
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
	egg, _ := errgroup.WithContext(ctx)
	egg.Go(func() error {
		_, err := io.Copy(local, workerConn)
		return err
	})
	egg.Go(func() error {
		_, err := io.Copy(workerConn, local)
		return err
	})

	return egg.Wait()
}

func (b *Bridge) readHandler(portal net.Conn, ctx context.Context) error {
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
			switch cmd := msg.(type) {
			case *protocol.ReqWorkCtl:
				log.Infof("get ReqWorkCtl")
				b.workingChan <- cmd.ProxyName
			case *protocol.Response:
				if cmd.Code == -1 {
					return errors.New(cmd.Message)
				}
			default:
				return errors.New("unknown command")
			}
		case <-ctx.Done():
			log.Infof("msgHandler exit with ctx done")
			return nil
		}

	}

}

func (b *Bridge) writeHandler(portal net.Conn, ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()
	for {
		select {
		case msg := <-b.msgWriteChan:
			err := protocol.WriteMsg(portal, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			log.Infof("witeHandler exit with ctx done")
			return nil
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
