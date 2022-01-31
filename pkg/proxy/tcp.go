package proxy

import (
	"breaker/pkg/netio"
	"breaker/pkg/protocol"
	"breaker/pkg/server"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type TcpProxy struct {
	Name string
	net.Listener
	WorkingChan chan net.Conn
	ctx         server.Context
}

func NewTcpProxy(name string, ctx server.Context) *TcpProxy {
	return &TcpProxy{
		Name:        name,
		ctx:         ctx,
		WorkingChan: make(chan net.Conn, 10),
	}
}
func (t *TcpProxy) Serve(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	t.Listener = listener
	go func() {
		defer func() {
			log.Infof("proxy:[%s] exist", t.Name)
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
				userconn.Close()
				return
			}
			go func() {
				clientWorkConn, err := t.GetWorkConn()
				if err != nil {
					log.Errorf("can not get work conn with err:[%+v]", err)
					userconn.Close()
					return
				}
				log.Infof("get worker connection:[%s]", clientWorkConn.RemoteAddr())
				netio.StartTunnel(clientWorkConn, userconn)
			}()

		}
	}()

	return nil
}
func (t *TcpProxy) GetWorkConn() (net.Conn, error) {
	// get a work connection from the chan
	for {
		select {
		case workConn := <-t.WorkingChan:
			log.Info("get work connection from chan")
			t.ctx.SetResponseMessage(&protocol.ReqWorkCtl{
				ProxyName: t.Name,
			})
			t.ctx.Send()
			return workConn, nil
		case <-time.After(time.Duration(5) * time.Second):
			return nil, errors.New("timeout trying to get work connection")
		}
	}
}

func (t *TcpProxy) Close() {
	t.Listener.Close()
	close(t.WorkingChan)
}
