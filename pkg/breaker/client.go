package breaker

import (
	"breaker/feature"
	"breaker/pkg/protocol"
	"breaker/plugin"
	"errors"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

var ErrClientStopped = errors.New("client stopped")

type Client struct {
	Conf *feature.BridgeConfig
	// Packer is the message packer, will be passed to Session.
	Packer Packer

	// Codec is the message codec, will be passed to Session.
	Codec Codec

	// OnSessionCreate is an event hook, will be invoked when Session's created.
	OnSessionCreate func(sess Session)

	// OnSessionClose is an event hook, will be invoked when Session's closed.
	OnSessionClose func(sess Session)

	socketReadBufferSize  int
	socketWriteBufferSize int
	readTimeout           time.Duration
	writeTimeout          time.Duration
	respQueueSize         int
	router                *Router
	stopped               chan struct{}
	writeAttemptTimes     int
	Session               *TcpSession
	FileServer            *plugin.FileServer
}

func NewClient(opts ...ClientOption) *Client {
	srv := &Client{

		stopped:           make(chan struct{}),
		Packer:            NewDefaultPacker(),
		Codec:             NewDefaultCodec(),
		router:            NewRouter(),
		respQueueSize:     QueueSize,
		writeAttemptTimes: DefaultWriteAttemptTimes,
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}
func (s *Client) Connect() error {
	conn, err := net.Dial("tcp", s.Conf.ServerAddr)
	if err != nil {
		return err
	}

	if s.socketReadBufferSize > 0 {
		if err := conn.(*net.TCPConn).SetReadBuffer(s.socketReadBufferSize); err != nil {
			return fmt.Errorf("conn set read buffer err: %s", err)
		}
	}
	if s.socketWriteBufferSize > 0 {
		if err := conn.(*net.TCPConn).SetWriteBuffer(s.socketWriteBufferSize); err != nil {
			return fmt.Errorf("conn set write buffer err: %s", err)
		}
	}
	log.Infof("start tcp client:" + s.Conf.ServerAddr)

	return s.login(conn)
}

func (s *Client) IsStopped() bool {
	select {
	case <-s.stopped:
		return true
	default:
		return false
	}
}

func (s *Client) handleSession() {

	if s.OnSessionCreate != nil {
		go s.OnSessionCreate(s.Session)
	}

	go s.Session.readInbound(s.router, s.readTimeout)               // start reading message packet from connection.
	go s.Session.writeOutbound(s.writeTimeout, s.writeAttemptTimes) // start writing message packet to connection.

	select {
	case <-s.Session.closed: // wait for Session finished.
	case <-s.stopped: // or the client is stopped.
	}

	if s.OnSessionClose != nil {
		go s.OnSessionClose(s.Session)
	}
}
func (s *Client) AddRoute(cmd protocol.Command, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	s.router.register(cmd, handler, middlewares...)
}

// Use registers global middlewares to the router.
func (s *Client) Use(middlewares ...MiddlewareFunc) {
	s.router.registerMiddleware(middlewares...)
}

// NotFoundHandler sets the not-found handler for router.
func (s *Client) NotFoundHandler(handler HandlerFunc) {
	s.router.setNotFoundHandler(handler)
}

// Stop stops client. Closing Listener and all connections.
func (s *Client) Stop() error {
	close(s.stopped)
	s.Session.Close()
	return nil
}

func (s *Client) login(c net.Conn) error {
	conn := &MasterConn{
		Conn: c,
	}
	session := NewTcpSession(conn,
		AsCodec(s.Codec),
		AsPacker(s.Packer),
		AsQueueSize(s.respQueueSize),
	)
	err := session.SendCmdSync(&protocol.NewMaster{})
	if err != nil {
		return err
	}
	cmdSync, err := session.ReadCmdSync()
	if err != nil {
		return err
	}
	resp, ok := cmdSync.(*protocol.NewMasterResp)
	if !ok {
		return fmt.Errorf("cmd is not NewMaster")
	}
	if resp.Error != "" {
		return fmt.Errorf(resp.Error)
	}

	sessionId := resp.SessionId

	log.Infof("login success,get Session id :[%s]", sessionId)

	session.SetID(sessionId)
	s.Session = session
	go s.handleSession()
	return nil
}

func (s *Client) Start() error {
	err := s.Connect()
	if err != nil {
		return err
	}
	//主动发送消息
	context := s.Session.AllocateContext()
	context.SetResponseMessage(&protocol.NewProxy{
		ProxyName:  s.Conf.ProxyName,
		RemotePort: s.Conf.RemotePort,
		TraceId:    s.Session.id.(string),
	})
	s.Session.Send(context)
	heartbeat := time.NewTicker(time.Duration(s.Conf.HeartbeatInterval) * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-heartbeat.C:
			context.SetResponseMessage(&protocol.Ping{})
			s.Session.Send(context)
		case <-s.Session.closed:
			for {
				log.Info("try to reconnect....")
				err := s.Connect()
				if err != nil {
					time.Sleep(time.Second * 5)
					continue
				}
				//主动发送消息
				context := s.Session.AllocateContext()
				context.SetResponseMessage(&protocol.NewProxy{
					ProxyName:  s.Conf.ProxyName,
					RemotePort: s.Conf.RemotePort,
					TraceId:    s.Session.id.(string),
				})
				s.Session.Send(context)
				break
			}
		case <-s.stopped:
			return nil
		}

	}
}
func (s *Client) CreateWorkerConn() (net.Conn, error) {
	//send worker
	sessionId := s.Session.ID().(string)
	workCmd := &protocol.NewWorkCtl{
		TraceID:   sessionId,
		ProxyName: s.Conf.ProxyName,
	}
	log.Infof("send message:[workCtl],Session id:[%s]", sessionId)
	log.Info("dial working server tcp:", s.Conf.ServerAddr)
	workerConn, err := net.Dial("tcp", s.Conf.ServerAddr)
	if err != nil {
		return nil, err
	}
	masterConn := &MasterConn{Conn: workerConn}
	err = masterConn.SendCmdSync(workCmd)
	if err != nil {
		return nil, err
	}
	cmdSync, err := masterConn.ReadCmdSync()
	if err != nil {
		return nil, err
	}
	workCtlResp, ok := cmdSync.(*protocol.NewWorkCtlResp)
	if !ok {
		return nil, errors.New("can't cast to NewWorkCtlResp")
	}
	if workCtlResp.Error != "" {
		return nil, errors.New(workCtlResp.Error)
	}
	return workerConn, nil
}

type ClientOption func(*Client)

func ClientConf(conf *feature.BridgeConfig) ClientOption {
	return func(client *Client) {
		client.Conf = conf
	}
}

func ClientPacker(pack Packer) ClientOption {
	return func(client *Client) {
		client.Packer = pack
	}
}

func ClientCodec(codec Codec) ClientOption {
	return func(client *Client) {
		client.Codec = codec
	}
}

func ClientOnSessionCreate(fn func(sess Session)) ClientOption {
	return func(client *Client) {
		client.OnSessionCreate = fn
	}
}
func ClientOnSessionClose(fn func(sess Session)) ClientOption {
	return func(client *Client) {
		client.OnSessionClose = fn
	}
}
func ClientSocketReadBufferSize(readSize int) ClientOption {
	return func(client *Client) {
		client.socketReadBufferSize = readSize
	}
}
func ClientSocketWriteBufferSize(writeSize int) ClientOption {
	return func(client *Client) {
		client.socketWriteBufferSize = writeSize
	}
}
func ClientReadTimeOut(readTimeout time.Duration) ClientOption {
	return func(s *Client) {
		s.readTimeout = readTimeout
	}

}

func ClientWriteTimeout(writeTimeout time.Duration) ClientOption {
	return func(s *Client) {
		s.writeTimeout = writeTimeout

	}
}

func ClientWriteAttemptTimes(writeAttemptTimes int) ClientOption {
	return func(s *Client) {
		s.writeAttemptTimes = writeAttemptTimes
	}
}
func ClientRespQueueSize(respQueueSize int) ClientOption {
	return func(s *Client) {
		s.respQueueSize = respQueueSize
	}
}
