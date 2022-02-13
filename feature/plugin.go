package feature

import "strings"

//PluginHttpProxy use for Portal
type PluginHttpProxy struct {
	ProxyPort int `ini:"proxy_port"`
}

func (b *PluginHttpProxy) OnInit() {
	if b.ProxyPort < 0 || b.ProxyPort > 65535 {
		panic("invalid local port[0-65535]")
	}
}

//PluginFileServer use for bridge
type PluginFileServer struct {
	FileLocation string `ini:"plugin_file_location"`
	Prefix       string `ini:"plugin_prefix"`
}

func (b *PluginFileServer) OnInit() {
	if b.Prefix == "" {
		b.Prefix = "/"
	}
	if !strings.HasPrefix(b.Prefix, "/") {
		b.Prefix = "/" + b.Prefix
	}
}
