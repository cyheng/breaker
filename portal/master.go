package portal

import (
	"breaker/pkg/protocol"
	"net"
	"sync"
	"sync/atomic"
)

type MasterManager struct {
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

	proxyLock sync.RWMutex
	once      sync.Once
}

func NewMaster(TrackID string, Conn net.Conn) *Master {

	return &Master{
		TrackID:   TrackID,
		Conn:      Conn,
		readChan:  make(chan interface{}, 20),
		writeChan: make(chan protocol.Command, 20),
	}
}

func (m *Master) Close() {
	m.once.Do(func() {
		m.Conn.Close()
		close(m.writeChan)
		close(m.readChan)
	})

}
