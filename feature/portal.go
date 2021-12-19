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
	ServerHost  string
	ServerPort  string
	listener    net.Listener
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
			log.Infof("get client connection [%s]", conn.RemoteAddr().String())
			go p.HandlerConn(conn, ctx)

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
func (p *Portal) HandlerConn(conn net.Conn, ctx context.Context) {
	defer conn.Close()
	for {
		msg, err := protocol.ReadMsg(conn)
		if err != nil {
			log.Info(err)
			return
		}

		switch msg.Type() {
		case protocol.TypeNewProxy:
			pxy := msg.(*protocol.NewProxy)
			err = p.onNewProxy(conn, pxy, ctx)
			break
		case protocol.TypeNewWorkCtl:
			workCtl := msg.(*protocol.WorkCtl)
			err = p.onNewWorkCtl(conn, workCtl, ctx)
			break
		case protocol.TypeCloseProxy:
			cmd := msg.(*protocol.CloseProxy)
			err = p.onCloseProxy(conn, cmd, ctx)
			break
		default:
			log.Debug("unknown command")
			err = errors.New("unknown command")
		}
		if err != nil {
			_ = protocol.WriteErrResponse(conn, err.Error())
		} else {
			_ = protocol.WriteSuccessResponse(conn)
		}
	}

}

//onNewProxy 收到TypeNewProxy指令之后，监听客户端发送过来的端口
func (p *Portal) onNewProxy(conn net.Conn, cmd *protocol.NewProxy, ctx context.Context) error {
	pxyName := cmd.ProxyName

	hostPort := net.JoinHostPort(p.ServerHost, strconv.Itoa(cmd.RemotePort))
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		log.Error("new Proxy error:", err)
		return err
	}
	log.Infof("newProxy with address:[%s]", hostPort)
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	if _, ok := p.RunningProxy[pxyName]; ok {
		log.Error("proxy already exist!")
		return errors.New("proxy already exist")
	}
	p.RunningProxy[pxyName] = listener
	return nil

}

func (p *Portal) onNewWorkCtl(clientWorkConn net.Conn, cmd *protocol.WorkCtl, ctx context.Context) error {
	log.Infof("get client working control:[%s],proxy:[%s]", clientWorkConn.RemoteAddr().String(), cmd.ProxyName)
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
			group, _ := errgroup.WithContext(ctx)
			group.Go(func() error {
				_, err := io.Copy(userconn, clientWorkConn)
				return err
			})
			group.Go(func() error {
				_, err := io.Copy(clientWorkConn, userconn)
				return err
			})
		}
	}()
	return nil
}

func (p *Portal) onCloseProxy(conn net.Conn, cmd *protocol.CloseProxy, ctx context.Context) error {
	log.Infof("close proxy:%s by client connection [%s]", cmd.ProxyName, conn.RemoteAddr().String())
	proxy, ok := p.RunningProxy[cmd.ProxyName]
	if !ok {
		return errors.New("proxy:" + cmd.ProxyName + " is not ready")
	}
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	err := proxy.Close()
	if err != nil {
		return err
	}
	delete(p.RunningProxy, cmd.ProxyName)
	return nil
}
