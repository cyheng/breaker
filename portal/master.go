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
	"runtime/debug"
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
	TrackID           string
	Conn              net.Conn
	WorkingConn       chan net.Conn
	readChan          chan interface{}
	writeChan         chan protocol.Command
	WorkingConnMaxCnt int
	//对用户访问的代理
	RunningProxy map[string]*proxy.TcpProxy

	proxyLock sync.RWMutex
	once      sync.Once
}

func NewMaster(TrackID string, Conn net.Conn) *Master {
	//TODO:配置working chan的数量
	WorkingConnMaxCnt := 10
	return &Master{
		TrackID:           TrackID,
		Conn:              Conn,
		WorkingConn:       make(chan net.Conn, WorkingConnMaxCnt),
		readChan:          make(chan interface{}, 20),
		writeChan:         make(chan protocol.Command, 20),
		WorkingConnMaxCnt: WorkingConnMaxCnt,
		RunningProxy:      make(map[string]*proxy.TcpProxy),
	}
}

//如何在Handler Message之中直到conn close?
func (m *Master) HandlerMessage(ctx context.Context) error {
	defer func() {
		log.Infof("exist handler message")
	}()
	ctx, cancel := context.WithCancel(ctx)
	egg, _ := errgroup.WithContext(ctx)
	egg.Go(func() error {
		return m.readMessage(ctx, cancel)
	})
	egg.Go(func() error {
		return m.handlerMessage(ctx)
	})
	egg.Go(func() error {
		return m.writeMessage(ctx)
	})
	//TODO:fix bug，readMessage阻塞
	err := egg.Wait()
	if err != nil {
		log.Errorf("%+v", err)
		cancel()
	}
	return err
}
func (m *Master) Close() {
	m.once.Do(func() {
		m.Conn.Close()
		close(m.WorkingConn)
		close(m.writeChan)
		close(m.readChan)
		m.CloseAllProxy()
	})

}

//onNewProxy 收到TypeNewProxy指令之后，监听客户端发送过来的端口
func (p *Master) onNewProxy(conn net.Conn, cmd *protocol.NewProxy, ctx context.Context) error {
	pxyName := cmd.ProxyName
	hostPort := net.JoinHostPort("0.0.0.0", strconv.Itoa(cmd.RemotePort))
	//TODO: conn close的时候，listener也要close
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
				listener.Close()
				return
			}
			group, _ := errgroup.WithContext(ctx)
			clientWorkConn, err := p.GetWorkConn()
			if err != nil {
				log.Errorf("can not get work conn with err:[%+v]", err)
				userconn.Close()
				continue
			}
			log.Infof("get worker connection:[%s]", clientWorkConn.RemoteAddr())
			//TODO:这个阻塞了
			group.Go(func() error {
				defer func() {
					userconn.Close()
				}()
				_, err := io.Copy(userconn, clientWorkConn)
				return err
			})
			group.Go(func() error {
				defer func() {
					clientWorkConn.Close()
				}()
				_, err := io.Copy(clientWorkConn, userconn)
				return err
			})
			err = group.Wait()
			if err != nil {
				log.Error(err)
			}

		}
	}()

	err2 := p.AddProxy(conn, pxyName, listener)
	if err2 != nil {
		return err2
	}
	res := protocol.Success()
	p.writeChan <- res
	return nil

}
func (p *Master) GetProxy(pxyName string) *proxy.TcpProxy {
	p.proxyLock.RLock()
	defer p.proxyLock.RUnlock()
	if pxy, ok := p.RunningProxy[pxyName]; ok {
		return pxy
	}

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
func (p *Master) CloseAllProxy() error {
	log.Infof("close all pxy")

	p.proxyLock.Lock()
	for _, tcpProxy := range p.RunningProxy {
		_ = tcpProxy.Close()
	}
	defer p.proxyLock.Unlock()

	return nil
}
func (p *Master) GetWorkConn() (net.Conn, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()

	// get a work connection from the chan
	for {
		select {
		case workConn := <-p.WorkingConn:
			log.Info("get work connection from chan")
			p.writeChan <- &protocol.ReqWorkCtl{}
			return workConn, nil
		case <-time.After(time.Duration(5) * time.Second):
			return nil, errors.New("timeout trying to get work connection")
		}
	}

}

func (m *Master) readMessage(ctx context.Context, cancel context.CancelFunc) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()

	for {
		//m.Conn.SetReadDeadline(time.Now().Add(time.Second * 5))

		msg, err := protocol.ReadMsg(m.Conn)
		//m.Conn.SetReadDeadline(time.Time{})

		if err != nil {
			log.Error(ctx.Value(TraceID), err)
			if err == protocol.ErrMsgFormat {
				continue
			}
			cancel()
			return err
		}
		m.readChan <- msg
	}

}

func (m *Master) handlerMessage(ctx context.Context) (err error) {
	for {
		select {
		case <-ctx.Done():
			m.Close()
			return ctx.Err()
		case msg := <-m.readChan:
			switch cmd := msg.(type) {
			case *protocol.NewProxy:
				err = m.onNewProxy(m.Conn, cmd, ctx)
			case *protocol.CloseProxy:
				err = m.onCloseProxy(cmd)
			default:
				log.Info("unknown command")
				err = errors.New("unknown command")
			}
			if err != nil {
				log.Error(ctx.Value(TraceID), err)
				cmd := protocol.Error(err.Error())
				m.writeChan <- cmd
			}

		}
	}
}

func (m *Master) writeMessage(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-m.writeChan:
			_ = protocol.WriteMsg(m.Conn, msg)
		}

	}
}
