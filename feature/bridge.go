package feature

import "strconv"

type BridgeConfig struct {
	LoggerConfig      `ini:"Logger"`
	PluginFileServer  `ini:"plugin_file_server"`
	ServerAddr        string `ini:"server_addr"`
	LocalPort         int    `ini:"local_port"`
	RemotePort        int    `ini:"remote_port"`
	ProxyName         string `ini:"proxy_name"`
	HeartbeatInterval int64  `ini:"heartbeat_interval" `
}

func (b *BridgeConfig) OnInit() {
	b.LoggerConfig.OnInit()
	b.PluginFileServer.OnInit()
	if b.ServerAddr == "" {
		panic("breaker address can not be empty")
	}
	if b.LocalPort < 0 || b.LocalPort > 65535 {
		panic("invalid local port[0-65535]")
	}
	if b.RemotePort < 0 || b.RemotePort > 65535 {
		panic("invalid remote port[0-65535]")
	}
	if b.HeartbeatInterval < 0 {
		panic("invalid HeartbeatInterval, can't less than 0")
	}
	if b.HeartbeatInterval == 0 {
		b.HeartbeatInterval = 5
	}
	if b.ProxyName == "" {
		b.ProxyName = b.ServerAddr + "_to_" + strconv.Itoa(b.LocalPort)
	}
}
