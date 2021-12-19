package feature

import (
	"breaker/pkg/protocol"
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

const FPortal = "portal"

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
		RunningProxy: make(map[string]net.Listener),
		proxyLock:    sync.Mutex{},
		WorkingConn:  make(chan WorkingConn),
	}
	res.ServerAddr = c.ServerAddr
	return res, nil
}

func init() {
	RegisterConfig(FPortal, &PortalConfig{})
}

type WorkingConn struct {
	ProxyName string
	Conn      net.Conn
}

//Portal implement the feature interface
type Portal struct {
	ServerAddr string
	//对用户访问的代理
	RunningProxy map[string]net.Listener
	//客户端发过来的连接
	WorkingConn chan WorkingConn
	proxyLock   sync.Mutex
}

func (p *Portal) Addr() string {
	return p.ServerAddr
}
func (p *Portal) Name() string {
	return FPortal
}
func (p *Portal) Stop(ctx context.Context) error {
	return nil
}
func (p *Portal) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", p.Addr())
	if err != nil {
		return err
	}
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
			log.Infof("get client connection [%s]", conn.RemoteAddr().String())
			go p.HandlerConn(conn)

		}
	})
	return egg.Wait()
}

// HandlerConn 判断是新建代理/关闭代理/心跳包
//客户端 待转发的端口 dial remote server
//新建代理(server):
//1.获取消息类型
//2.NewProxy:获取客户端发过来的端口remote_port和proxy_name,worker
// 			 服务端proxyListener :=net.Listen(remote_port),
//
////3.TypeNewWorkCtl:获取客户端的workConn(type:NewWorkConn)
//	io.copy userconn和WorkConn 的connect,clientConn net.Dial(client_addr) io.copy(connect,clientConn)
//4.CloseProxy:获取proxy_name,关闭对应的proxy
//5.Ping:返回pong
//新建代理(client):
//1. control:=connect to server(server_addr:port)
//2. connect to localport(需要代理的port,proxy conn)
//3. worker:=connect to server(server_addr:port)
//4. control 发送NewProxy:remote_port和proxy_name
//5. copy(worker,proxy conn)
//6. control 发送NewProxyWorker
func (p *Portal) HandlerConn(conn net.Conn) {
	defer conn.Close()
	msg, err := protocol.ReadMsg(conn)
	if err != nil {
		log.Info(err)
		return
	}
	switch msg.Type() {
	case protocol.TypeNewProxy:
		pxy := msg.(*protocol.NewProxy)
		p.onNewProxy(pxy)
		break
	case protocol.TypeNewWorkCtl:
		workCtl := msg.(*protocol.WorkCtl)
		p.onNewWorkCtl(conn, workCtl)
		break
	case protocol.TypeCloseProxy:
		cmd := msg.(*protocol.CloseProxy)
		p.onCloseProxy(cmd)
	default:
		log.Debug("unknown command")
		return
	}

}

//onNewProxy 收到TypeNewProxy指令之后，监听客户端发送过来的端口
func (p *Portal) onNewProxy(cmd *protocol.NewProxy) error {
	pxyName := cmd.ProxyName

	hostPort := net.JoinHostPort(p.ServerAddr, strconv.Itoa(cmd.RemotePort))
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		log.Error("new Proxy error:", err)
		return err
	}
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	if _, ok := p.RunningProxy[pxyName]; ok {
		log.Error("proxy already exist!")
		return errors.New("proxy already exist")
	}
	p.RunningProxy[pxyName] = listener
	return nil

}

func (p *Portal) onNewWorkCtl(clientWorkConn net.Conn, cmd *protocol.WorkCtl) error {
	proxy, ok := p.RunningProxy[cmd.ProxyName]
	if !ok {
		return errors.New("proxy:" + cmd.ProxyName + " is not ready")
	}
	go func() {
		for {
			userconn, err := proxy.Accept()
			var tempDelay time.Duration // how long to sleep on accept failure
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
				log.Infof("met proxy accept error: %s", err)

				return
			}

			go io.Copy(userconn, clientWorkConn)
			go io.Copy(clientWorkConn, userconn)
		}
	}()
	return nil
}

func (p *Portal) onCloseProxy(cmd *protocol.CloseProxy) error {
	proxy, ok := p.RunningProxy[cmd.ProxyName]
	if !ok {
		return errors.New("proxy:" + cmd.ProxyName + " is not ready")
	}
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	proxy.Close()
	delete(p.RunningProxy, cmd.ProxyName)
	return nil
}
