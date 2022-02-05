package main

import (
	"breaker/feature"
	"breaker/pkg/breaker"
	"breaker/pkg/protocol"
	"breaker/pkg/proxy"
	"breaker/portal"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	version = "1.0.0"
)

var (
	cfgFile     string
	showVersion bool
)

var cmdRoot = &cobra.Command{
	Use: "portal",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println(version)
		}
		conf := &feature.PortalConfig{}
		err := feature.LoadFromFile(cfgFile, conf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)

		}
		conf.OnInit()
		srv := NewPortalService(conf)
		go func() {
			if err := srv.Serve(conf.ServerAddr); err != nil {
				log.Error(err)
			}
		}()
		// 后台运行
		{
			osSignals := make(chan os.Signal, 1)
			signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
			<-osSignals
			if err := srv.Stop(); err != nil {
				log.Errorf("portal  stopped err: %s", err)
			}
		}
	},
}

func init() {
	cmdRoot.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.ini", "config file of breaker")
	cmdRoot.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of breaker")

}

func NewPortalService(conf *feature.PortalConfig) *breaker.Server {
	srv := breaker.NewServer()
	masterManager := portal.NewMasterManager()
	pm := &proxy.ProxyManager{
		RunningProxy: make(map[string]*proxy.TcpProxy),
	}
	srv.OnSessionClose = func(sess breaker.Session) {
		sessid := sess.ID().(string)
		err := masterManager.DeleteMaster(sessid)
		if err == nil {
			log.Infof("close master with session id:[%s]", sessid)
		}
		err = pm.DeleteProxy(sessid)
		if err == nil {
			log.Infof("close proxy with session id:[%s]", sessid)
		}
	}
	srv.Use(breaker.RecoverMiddleware())
	srv.AddRoute(&protocol.NewMaster{}, func(ctx breaker.Context) {
		conn := ctx.Conn()
		sessid := ctx.Session().ID().(string)
		master := portal.NewMaster(sessid, conn)
		masterManager.AddMaster(master)
		log.Infof("new master with session id :[%s]", sessid)
		ctx.SetResponseMessage(&protocol.NewMasterResp{
			SessionId: sessid,
		})
	})
	srv.AddRoute(&protocol.Ping{}, func(ctx breaker.Context) {
		log.Info("get ping from client")
		ctx.SetResponseMessage(&protocol.Pong{})
	})
	//从客户端中获取Working conn
	srv.AddRoute(&protocol.ReqWorkCtlResp{}, func(ctx breaker.Context) {

	})
	srv.AddRoute(&protocol.NewWorkCtl{}, func(ctx breaker.Context) {
		cmd := ctx.Request().(*protocol.NewWorkCtl)
		clientWorkConn := ctx.Conn()
		log.Infof("get client working control:[%s],trace id:[%s]", clientWorkConn.RemoteAddr().String(), cmd.TraceID)
		resp := &protocol.NewWorkCtlResp{}
		pxy, ok := pm.GetProxy(cmd.TraceID)
		if !ok {
			log.Errorf("working control:[%s] error:proxy not found", clientWorkConn.RemoteAddr().String())
			resp.Error = fmt.Sprintf("working control:[%s] error:proxy not found", clientWorkConn.RemoteAddr().String())
			ctx.SetResponseMessage(resp).SendSync()
			return
		}
		select {
		case pxy.WorkingChan <- clientWorkConn:
			log.Info("new work connection registered")

			ctx.SetResponseMessage(resp).SendSync()
			return
		default:
			log.Errorf("work connection pool is full, discarding")
			resp.Error = fmt.Sprintf("work connection pool is full, discarding")
			ctx.SetResponseMessage(resp).SendSync()
			return
		}
	}, func(next breaker.HandlerFunc) breaker.HandlerFunc {
		return func(ctx breaker.Context) {
			next(ctx)
			ctx.Session().Close()
		}
	})
	srv.AddRoute(&protocol.NewProxy{}, func(ctx breaker.Context) {
		cmd := ctx.Request().(*protocol.NewProxy)
		sessid := cmd.TraceId

		pxyName := cmd.ProxyName
		hostPort := net.JoinHostPort("0.0.0.0", strconv.Itoa(cmd.RemotePort))
		pxy := proxy.NewTcpProxy(pxyName, ctx)
		err := pxy.Serve(hostPort)
		resp := &protocol.NewProxyResp{
			ProxyName: pxyName,
		}
		if err != nil {
			log.Error("new Proxy error:", err)
			resp.Error = "new Proxy error:%+v" + err.Error()
			ctx.SetResponseMessage(resp)
			return
		}
		log.Infof("newProxy with address:[%s],session id:[%s]", hostPort, sessid)
		err = pm.AddProxy(sessid, pxy)
		if err != nil {
			resp.Error = "add Proxy error:%+v" + err.Error()
			ctx.SetResponseMessage(resp)
			return
		}
		ctx.SetResponseMessage(resp)
	})
	srv.AddRoute(&protocol.CloseProxy{}, func(ctx breaker.Context) {
		cmd := ctx.Request().(*protocol.CloseProxy)
		sessid := ctx.Session().ID().(string)
		log.Infof("close pxy:%s  ", cmd.ProxyName)
		err := pm.DeleteProxy(sessid)
		resp := &protocol.CloseProxyResp{}
		if err != nil {
			resp.Error = "close Proxy error:%+v" + err.Error()
			ctx.SetResponseMessage(resp)
			return
		}
		ctx.SetResponseMessage(resp)
	})
	return srv

}

func main() {
	err := cmdRoot.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
