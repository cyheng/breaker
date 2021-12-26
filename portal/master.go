package portal

import (
	"breaker/pkg/protocol"
	"breaker/pkg/proxy"
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

type MasterManager struct {
	masterByTrackID map[string]*Master

	mu sync.RWMutex
}

func NewMasterManager() *MasterManager {
	return &MasterManager{
		masterByTrackID: make(map[string]*Master),
	}
}

func (m *MasterManager) AddMaster(master *Master) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.masterByTrackID[master.TrackID] = master
}

func (m *MasterManager) DeleteMaster(traceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.masterByTrackID, traceID)
}

func (m *MasterManager) GetMaster(traceID string) *Master {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.masterByTrackID[traceID]
}

type Master struct {
	TrackID     string
	Conn        net.Conn
	WorkingConn chan net.Conn
	//对用户访问的代理
	RunningProxy map[string]*proxy.TcpProxy

	proxyLock sync.Mutex
}

func NewMaster(TrackID string, Conn net.Conn) *Master {
	return &Master{
		TrackID: TrackID,
		Conn:    Conn,
		//TODO:配置working chan的数量
		WorkingConn:  make(chan net.Conn, 10),
		RunningProxy: make(map[string]*proxy.TcpProxy),
	}
}
func (m *Master) HandlerMessage(ctx context.Context) {
	defer func() {
		log.Infof("exist handler message")
	}()
	for {
		select {
		case <-ctx.Done():
			m.Close()
			return
		default:
			msg, err := protocol.ReadMsg(m.Conn)
			if err != nil {
				log.Error(ctx.Value(TraceID), err)
				if err == protocol.ErrMsgFormat {
					continue
				}
				return
			}
			switch cmd := msg.(type) {
			case *protocol.NewProxy:
				err = m.onNewProxy(m.Conn, cmd, ctx)
				break
			case *protocol.CloseProxy:
				err = m.onCloseProxy(cmd)
				break
			default:
				log.Debug("unknown command")
				err = errors.New("unknown command")
			}
			if err != nil {
				log.Error(ctx.Value(TraceID), err)
				_ = protocol.WriteErrResponse(m.Conn, err.Error())
			}
		}
	}
}
func (m *Master) Close() {
	m.Conn.Close()
	close(m.WorkingConn)
}

//onNewProxy 收到TypeNewProxy指令之后，监听客户端发送过来的端口
func (p *Master) onNewProxy(conn net.Conn, cmd *protocol.NewProxy, ctx context.Context) error {
	pxyName := cmd.ProxyName

	hostPort := net.JoinHostPort("0.0.0.0", strconv.Itoa(cmd.RemotePort))
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		log.Error("new Proxy error:", err)
		return err
	}
	log.Infof("newProxy with address:[%s]", hostPort)

	go func() {
		defer func() {
			log.Infof("proxy:[%s] exist", cmd.ProxyName)
		}()
		for {
			userconn, err := listener.Accept()
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
				log.Infof("met pxy accept error: %s", err)

				return
			}
			group, _ := errgroup.WithContext(ctx)
			clientWorkConn := <-p.WorkingConn
			log.Infof("get worker connection:[%s]", clientWorkConn.RemoteAddr())
			group.Go(func() error {
				_, err := io.Copy(userconn, clientWorkConn)
				return err
			})
			group.Go(func() error {
				_, err := io.Copy(clientWorkConn, userconn)
				return err
			})
			err = group.Wait()
			if err != nil {
				//TODO:判断WorkingConn是否关闭
				//Working Conn回归到pool中
				//p.WorkingConn <- clientWorkConn
				log.Error(err)
			}

		}
	}()

	err2 := p.AddProxy(conn, pxyName, listener)
	if err2 != nil {
		return err2
	}
	_ = protocol.WriteSuccessResponse(p.Conn)
	return nil

}

func (p *Master) AddProxy(conn net.Conn, pxyName string, listener net.Listener) error {
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	if _, ok := p.RunningProxy[pxyName]; ok {
		log.Error("proxy already exist!")
		return errors.New("proxy already exist")
	}

	p.RunningProxy[pxyName] = proxy.NewTcpProxy(pxyName, listener, conn)
	return nil
}

func (p *Master) onCloseProxy(cmd *protocol.CloseProxy) error {
	log.Infof("close pxy:%s  ", cmd.ProxyName)
	pxy, ok := p.RunningProxy[cmd.ProxyName]
	if !ok {
		return errors.New("pxy:" + cmd.ProxyName + " is not ready")
	}
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	err := pxy.Close()
	if err != nil {
		return err
	}
	delete(p.RunningProxy, cmd.ProxyName)
	return nil
}

//在连接断开的时候，关闭代理
func (p *Master) CloseProxy(conn net.Conn) error {
	panic("not ready")
	return nil
}
