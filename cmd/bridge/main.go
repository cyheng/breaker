package main

import (
	"breaker/feature"
	"breaker/pkg/breaker"
	"breaker/pkg/netio"
	"breaker/pkg/protocol"
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
	Use: "bridge",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Println(version)
		}
		conf := &feature.BridgeConfig{}
		err := feature.LoadFromFile(cfgFile, conf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		conf.OnInit()
		cli := NewBridge(conf)
		err = cli.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// 后台运行
		{
			osSignals := make(chan os.Signal, 1)
			signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
			<-osSignals
			if err := cli.Stop(); err != nil {
				log.Errorf("bridge stopped err: %s", err)
			}
		}
	},
}

func init() {
	cmdRoot.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.ini", "config file of breaker")
	cmdRoot.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "version of breaker")

}

func NewBridge(conf *feature.BridgeConfig) *breaker.Client {
	cli := breaker.NewClient(
		breaker.ClientConf(conf),
	)
	cli.Use(breaker.RecoverMiddleware())

	cli.AddRoute(&protocol.NewProxyResp{}, func(ctx breaker.Context) {
		log.Infof("get message NewProxyResp,session id :[%s]", ctx.Session().ID())
		cmd := ctx.Request().(*protocol.NewProxyResp)
		ctx.SetRedirectMessage(&protocol.ReqWorkCtl{
			ProxyName: cmd.ProxyName,
		})
	})
	cli.AddRoute(&protocol.CloseProxyResp{}, func(ctx breaker.Context) {

	})
	cli.AddRoute(&protocol.Pong{}, func(ctx breaker.Context) {
		log.Info("get pong from server")

	})
	cli.AddRoute(&protocol.ReqWorkCtl{}, func(ctx breaker.Context) {
		workerConn, err := cli.CreateWorkerConn()
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		addr := net.JoinHostPort("0.0.0.0", strconv.Itoa(cli.Conf.LocalPort))
		log.Tracef("dial local tcp:[%s]", addr)
		local, err := net.Dial("tcp", addr)
		if err != nil {
			workerConn.Close()
			log.Errorf(err.Error())
			return
		}
		go netio.StartTunnel(workerConn, local)
	})

	cli.AddRoute(&protocol.NewWorkCtlResp{}, func(ctx breaker.Context) {

	})
	return cli
}

func main() {
	err := cmdRoot.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
