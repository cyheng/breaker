package feature

import (
	"strconv"
)

const FBridge = "bridge"

func init() {
	RegisterConfig(FBridge, &BridgeConfig{})
}

func (b *BridgeConfig) OnInit() {

	if b.ServerAddr == "" {
		panic("server address can not be empty")
	}
	if b.LocalPort < 0 || b.LocalPort > 65535 {
		panic("invalid local port[0-65535]")
	}
	if b.RemotePort < 0 || b.RemotePort > 65535 {
		panic("invalid local port[0-65535]")
	}
	if b.ProxyName == "" {
		b.ProxyName = b.ServerAddr + "_to_" + strconv.Itoa(b.LocalPort)
	}
}
func (b *BridgeConfig) ServiceName() string {
	return FBridge
}
