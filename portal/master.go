package portal

import (
	"breaker/pkg/protocol"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type MasterManager struct {
	masterByTrackID  sync.Map
	masterNum        int64
	HeartbeatTimeout int64
}

func NewMasterManager() *MasterManager {
	return &MasterManager{
		masterNum:        0,
		masterByTrackID:  sync.Map{},
		HeartbeatTimeout: 90,
	}
}

func (m *MasterManager) AddMaster(master *Master) {
	m.masterByTrackID.Store(master.TrackID, master)
	atomic.AddInt64(&m.masterNum, 1)

}

func (m *MasterManager) DeleteMaster(traceID string) error {
	v, ok := m.masterByTrackID.Load(traceID)
	if !ok {
		return fmt.Errorf("session id:[%s] not found in master", traceID)
	}
	master := v.(*Master)
	master.Close()
	m.masterByTrackID.Delete(traceID)
	atomic.AddInt64(&m.masterNum, -1)
	return nil
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

func (m *MasterManager) CheckConn() {
	heartbeat := time.NewTicker(time.Second * 30)
	defer heartbeat.Stop()
	for {
		select {
		case <-heartbeat.C:
			m.masterByTrackID.Range(func(key, value interface{}) bool {
				master := value.(*Master)
				if time.Since(master.LastPingTime) > time.Duration(m.HeartbeatTimeout)*time.Second {
					master.Close()
				}
				return true
			})
		}
	}

}

//Master 客户端和服务端的
type Master struct {
	TrackID      string
	Conn         net.Conn
	readChan     chan interface{}
	writeChan    chan protocol.Command
	once         sync.Once
	LastPingTime time.Time
}

func NewMaster(TrackID string, Conn net.Conn) *Master {

	return &Master{
		TrackID:      TrackID,
		Conn:         Conn,
		readChan:     make(chan interface{}, 20),
		writeChan:    make(chan protocol.Command, 20),
		LastPingTime: time.Now(),
	}
}

func (m *Master) Close() {
	m.once.Do(func() {
		m.Conn.Close()
		close(m.writeChan)
		close(m.readChan)
	})

}
