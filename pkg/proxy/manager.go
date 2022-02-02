package proxy

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
)

type ProxyManager struct {
	//对用户访问的代理
	RunningProxy map[string]*TcpProxy

	proxyLock sync.RWMutex
}

func (p *ProxyManager) AddProxy(sessid string, t *TcpProxy) error {
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	if _, ok := p.RunningProxy[t.Name]; ok {
		log.Error("proxy already exist!")
		return errors.New("proxy already exist")
	}

	p.RunningProxy[sessid] = t
	return nil
}

func (p *ProxyManager) DeleteProxy(sessid string) error {
	p.proxyLock.Lock()
	defer p.proxyLock.Unlock()
	pxy, ok := p.RunningProxy[sessid]
	if !ok {
		return errors.New("pxy:" + sessid + " is not ready")
	}
	pxy.Close()
	delete(p.RunningProxy, sessid)
	return nil
}
func (p *ProxyManager) GetProxy(sessid string) (*TcpProxy, bool) {
	p.proxyLock.RLock()
	defer p.proxyLock.RUnlock()
	if pxy, ok := p.RunningProxy[sessid]; ok {
		return pxy, ok
	}

	return nil, false
}
