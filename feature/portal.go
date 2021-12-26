package feature

import (
	"breaker/pkg/protocol"
	"breaker/pkg/uuid"
	"breaker/portal"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"net"
	"runtime/debug"
	"time"
)

const (
	FPortal = "portal"
)

type PortalConfig struct {
	ServerAddr string `ini:"server_addr"`
}

func (c *PortalConfig) OnInit() {
	if c.ServerAddr == "" {
		c.ServerAddr = "0.0.0.0:80"
	}
}

func (c *PortalConfig) NewFeature() (Feature, error) {
	res := &Portal{
		masterManager: portal.NewMasterManager(),
	}
	res.ServerAddr = c.ServerAddr
	host, port, err := net.SplitHostPort(c.ServerAddr)
	if err != nil {
		return nil, err
	}
	res.ServerHost = host
	res.ServerPort = port
	return res, nil
}

func init() {
	RegisterConfig(FPortal, &PortalConfig{})
}

//Portal implement the feature interface
type Portal struct {
	ServerAddr string

	ServerHost    string
	ServerPort    string
	listener      net.Listener
	masterManager *portal.MasterManager
}

func (p *Portal) Addr() string {
	return p.ServerAddr
}
func (p *Portal) Name() string {
	return FPortal
}
func (p *Portal) Stop(ctx context.Context) error {
	if p.listener != nil {
		p.listener.Close()
	}
	return nil
}
func (p *Portal) Start(ctx context.Context) error {

	listener, err := net.Listen("tcp", p.Addr())
	if err != nil {
		return err
	}
	p.listener = listener
	log.Printf("%v listening TCP on %v", p.Name(), p.Addr())
	egg, ctx := errgroup.WithContext(ctx)

	egg.Go(func() error {
		var tempDelay time.Duration // how long to sleep on accept failure
		for {
			conn, err := listener.Accept()
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					log.Infof("met temporary error: %s, sleep for %s ...", err, tempDelay)
					time.Sleep(tempDelay)
					continue
				}

				return err
			}
			ctx = context.WithValue(ctx, portal.TraceID, uuid.New().String())
			log.Infof("get client connection [%s]", conn.RemoteAddr().String())
			go p.HandlerConn(conn, ctx)

		}
	})
	return egg.Wait()
}

// HandlerConn 只会处理两个命令，control和workCtl
// master 负责portal和bridge之间的沟通, 如：新建代理/关闭代理/心跳包
// workCtl 注册 客户端的coon,用于流量转发
func (p *Portal) HandlerConn(conn net.Conn, ctx context.Context) {
	//defer conn.Close()
	msg, err := protocol.ReadMsg(conn)
	if err != nil {
		log.Error(ctx.Value(portal.TraceID), err)
		conn.Close()
		return
	}
	switch cmd := msg.(type) {

	case *protocol.NewMaster:
		err = p.onNewMaster(conn, cmd, ctx)
		break
	case *protocol.WorkCtl:
		err = p.onNewWorkCtl(conn, cmd)
		break
	default:
		log.Debug("unknown command")
		err = errors.New("unknown command")
	}
	if err != nil {
		conn.Close()
	}
}

// 新建一个goroutine 负责和bridge沟通
func (p *Portal) onNewMaster(conn net.Conn, cmd *protocol.NewMaster, ctx context.Context) error {
	traceID := ctx.Value(portal.TraceID).(string)
	master := portal.NewMaster(traceID, conn)
	log.Infof("add master[%s]", traceID)
	p.masterManager.AddMaster(master)

	go master.HandlerMessage(ctx)
	return protocol.WriteSuccessResponseWithData(conn, traceID)
}
func (p *Portal) onNewWorkCtl(clientWorkConn net.Conn, cmd *protocol.WorkCtl) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()

	log.Infof("get client working control:[%s],trace id:[%s]", clientWorkConn.RemoteAddr().String(), cmd.TraceID)
	master := p.masterManager.GetMaster(cmd.TraceID)
	if master == nil {
		log.Errorf(" working control:[%s] error:master not found", clientWorkConn.RemoteAddr().String())
		_ = protocol.WriteErrResponse(clientWorkConn, "master not found")
		return errors.New("master not found")
	}

	select {
	case master.WorkingConn <- clientWorkConn:
		log.Debug("new work connection registered")
		_ = protocol.WriteSuccessResponse(clientWorkConn)
		return nil
	default:
		log.Debug("work connection pool is full, discarding")
		_ = protocol.WriteErrResponse(clientWorkConn, "work connection pool is full, discarding")
		return fmt.Errorf("work connection pool is full, discarding")
	}

}
