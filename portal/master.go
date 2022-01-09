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
	"sync/atomic"
	"time"
)

type MasterManager struct {
	//todo:使用sync.map
	masterByTrackID sync.Map
	masterNum       int64
}

func NewMasterManager() *MasterManager {
	return &MasterManager{
		masterNum:       0,
		masterByTrackID: sync.Map{},
	}
}

func (m *MasterManager) AddMaster(master *Master) {
	m.masterByTrackID.Store(master.TrackID, master)
	atomic.AddInt64(&m.masterNum, 1)

}

func (m *MasterManager) DeleteMaster(traceID string) {
	m.masterByTrackID.Delete(traceID)
	atomic.AddInt64(&m.masterNum, -1)
}
func (s *MasterManager) GetMasterNum() int64 {
	return atomic.LoadInt64(&s.masterNum)
}
func (m *MasterManager) GetMaster(traceID string) (*Master, bool) {
	v, ok := m.masterByTrackID.Load(traceID)
	if !ok {
		return nil, ok
	}
	return v.(*Master), ok
}
func (m *MasterManager) Range(f func(traceId, master interface{}) bool) {
	m.masterByTrackID.Range(f)
}

//Master 客户端和服务端的
type Master struct {
	TrackID string
	Conn    net.Conn

	readChan          chan interface{}
	writeChan         chan protocol.Command
	WorkingConnMaxCnt int

	//对用户访问的代理
	RunningProxy map[string]*proxy.TcpProxy

	proxyLock sync.RWMutex
	once      sync.Once
}

func NewMaster(TrackID string, Conn net.Conn) *Master {

	return &Master{
		TrackID:      TrackID,
		Conn:         Conn,
		readChan:     make(chan interface{}, 20),
		writeChan:    make(chan protocol.Command, 20),
		RunningProxy: make(map[string]*proxy.TcpProxy),
	}
}

//如何在Handler Message之中直到conn close?
func (m *Master) HandlerMessage(ctx context.Context, closeFn func(m *Master)) error {
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

	err := egg.Wait()
	if err != nil {
		log.Errorf("%+v", err)
		cancel()
		closeFn(m)
	}
	return err
}
func (m *Master) Close() {
	m.once.Do(func() {
		m.Conn.Close()
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
	err2 := p.AddProxy(conn, pxyName, listener)
	if err2 != nil {
		return err2
	}
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
			clientWorkConn, err := p.GetWorkConn(cmd.ProxyName)
			if err != nil {
				log.Errorf("can not get work conn with err:[%+v]", err)
				userconn.Close()
				continue
			}
			log.Infof("get worker connection:[%s]", clientWorkConn.RemoteAddr())
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

	res := protocol.Success()
	p.writeChan <- res
	return nil

}
func (p *Master) GetProxy(pxyName string) (*proxy.TcpProxy, bool) {
	p.proxyLock.RLock()
	defer p.proxyLock.RUnlock()
	if pxy, ok := p.RunningProxy[pxyName]; ok {
		return pxy, ok
	}

	return nil, false
}
func (p *Master) AddProxy(conn net.Conn, pxyName string, listener net.Listener) error {
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	if _, ok := p.RunningProxy[pxyName]; ok {
		log.Error("proxy already exist!")
		return errors.New("proxy already exist")
	}

	p.RunningProxy[pxyName] = proxy.NewTcpProxy(pxyName, listener)
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
	pxy.Close()

	delete(p.RunningProxy, cmd.ProxyName)
	return nil
}
func (p *Master) CloseAllProxy() error {
	log.Infof("close all pxy")

	p.proxyLock.Lock()
	for _, tcpProxy := range p.RunningProxy {
		tcpProxy.Close()
	}
	defer p.proxyLock.Unlock()

	return nil
}

//应该在proxy manger 中
func (p *Master) GetWorkConn(pxyName string) (net.Conn, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error("panic error: %v", err)
			log.Error(string(debug.Stack()))
		}
	}()
	pxy, ok := p.GetProxy(pxyName)
	if !ok {
		return nil, errors.New("proxy not found")
	}

	// get a work connection from the chan
	for {
		select {
		case workConn := <-pxy.WorkingChan:
			log.Info("get work connection from chan")
			p.writeChan <- &protocol.ReqWorkCtl{
				ProxyName: pxyName,
			}
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

		msg, err := protocol.ReadMsg(m.Conn)
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
