package main

import (
	"breaker/pkg/protocol"
	"breaker/pkg/proxy"
	"breaker/pkg/server"
	"breaker/portal"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

var masterManager *portal.MasterManager

type ProxyManager struct {
	//对用户访问的代理
	RunningProxy map[string]*proxy.TcpProxy

	proxyLock sync.RWMutex
}

func (p *ProxyManager) AddProxy(sessid string, t *proxy.TcpProxy) error {
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	if _, ok := p.RunningProxy[t.Name]; ok {
		log.Error("proxy already exist!")
		return errors.New("proxy already exist")
	}

	p.RunningProxy[sessid] = t
	return nil
}

func (p *ProxyManager) DeleteProxy(sessid string) error {
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	pxy, ok := p.RunningProxy[sessid]
	if !ok {
		return errors.New("pxy:" + sessid + " is not ready")
	}
	pxy.Close()
	delete(p.RunningProxy, sessid)
	return nil
}
func (p *ProxyManager) GetProxy(sessid string) (*proxy.TcpProxy, bool) {
	p.proxyLock.RLock()
	defer p.proxyLock.RUnlock()
	if pxy, ok := p.RunningProxy[sessid]; ok {
		return pxy, ok
	}

	return nil, false
}

func NewMasterServer() {
	srv := server.NewServer()
	masterManager := portal.NewMasterManager()
	pm := &ProxyManager{
		RunningProxy: make(map[string]*proxy.TcpProxy),
	}

	srv.Use(server.RecoverMiddleware())
	srv.AddRoute(protocol.TypeNewMaster, func(ctx server.Context) {
		conn := ctx.Conn()
		traceID := ctx.Session().ID().(string)
		master := portal.NewMaster(traceID, conn)
		masterManager.AddMaster(master)
		ctx.SetResponseMessage(protocol.SuccessWithData(traceID))
	})
	srv.AddRoute(protocol.TypeNewWorkCtl, func(ctx server.Context) {
		cmd := ctx.Request().(*protocol.WorkCtl)
		clientWorkConn := ctx.Conn()
		log.Infof("get client working control:[%s],trace id:[%s]", clientWorkConn.RemoteAddr().String(), cmd.TraceID)

		pxy, ok := pm.GetProxy(cmd.TraceID)
		if !ok {
			log.Errorf(" working control:[%s] error:proxy not found", clientWorkConn.RemoteAddr().String())
			ctx.SetResponseMessage(protocol.Error("proxy not found")).SendSync()
			return
		}
		select {
		case pxy.WorkingChan <- clientWorkConn:
			log.Info("new work connection registered")
			ctx.SetResponseMessage(protocol.Success()).SendSync()
			return
		default:
			log.Errorf("work connection pool is full, discarding")
			ctx.SetResponseMessage(protocol.Error("work connection pool is full, discarding")).SendSync()
			return
		}
	}, func(next server.HandlerFunc) server.HandlerFunc {
		return func(ctx server.Context) {
			next(ctx)
			ctx.Session().Close()
		}
	})
	srv.AddRoute(protocol.TypeNewProxy, func(ctx server.Context) {
		cmd := ctx.Request().(*protocol.NewProxy)
		pxyName := cmd.ProxyName
		hostPort := net.JoinHostPort("0.0.0.0", strconv.Itoa(cmd.RemotePort))
		pxy := proxy.NewTcpProxy(pxyName, ctx)
		err := pxy.Serve(hostPort)
		if err != nil {
			log.Error("new Proxy error:", err)
			ctx.SetResponseMessage(protocol.Error("new Proxy error:%+v" + err.Error()))
			return
		}
		log.Infof("newProxy with address:[%s]", hostPort)
		sessid := ctx.Session().ID().(string)
		err = pm.AddProxy(sessid, pxy)
		if err != nil {
			ctx.SetResponseMessage(protocol.Error("add Proxy error:%+v" + err.Error()))
			return
		}

		ctx.SetResponseMessage(protocol.Success())

	})
	srv.AddRoute(protocol.TypeCloseProxy, func(ctx server.Context) {
		cmd := ctx.Request().(*protocol.CloseProxy)
		sessid := ctx.Session().ID().(string)
		log.Infof("close pxy:%s  ", cmd.ProxyName)
		err := pm.DeleteProxy(sessid)
		if err != nil {
			ctx.SetResponseMessage(protocol.Error("new Proxy error:%+v" + err.Error()))
			return
		}
		ctx.SetResponseMessage(protocol.Success())
	})

	go func() {
		if err := srv.Serve("0.0.0.0:8899"); err != nil {
			log.Error(err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	if err := srv.Stop(); err != nil {
		log.Errorf("server stopped err: %s", err)
	}
}

func main() {
	NewMasterServer()
}
