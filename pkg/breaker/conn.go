package breaker

import (
	"breaker/pkg/protocol"
	"net"
	"time"
)

//MasterConn use for client,each breaker client has its owen MasterConn
type MasterConn struct {
	Conn net.Conn
}

func (m *MasterConn) Read(b []byte) (n int, err error) {
	return m.Conn.Read(b)
}

func (m *MasterConn) Write(b []byte) (n int, err error) {
	return m.Conn.Write(b)
}

func (m *MasterConn) Close() error {
	return m.Conn.Close()
}

func (m *MasterConn) LocalAddr() net.Addr {
	return m.Conn.LocalAddr()
}

func (m *MasterConn) RemoteAddr() net.Addr {
	return m.Conn.LocalAddr()
}

func (m *MasterConn) SetDeadline(t time.Time) error {
	return m.Conn.SetReadDeadline(t)
}

func (m *MasterConn) SetReadDeadline(t time.Time) error {
	return m.Conn.SetReadDeadline(t)
}

func (m *MasterConn) SetWriteDeadline(t time.Time) error {
	return m.Conn.SetWriteDeadline(t)
}
func (m *MasterConn) SendCmdSync(command protocol.Command) error {
	return protocol.WriteMsg(m, command)
}
func (m *MasterConn) ReadCmdSync() (protocol.Command, error) {
	return protocol.ReadMsg(m)
}
