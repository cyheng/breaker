package feature

//PluginHttpProxy use for Portal

type PluginHttpProxy struct {
	ProxyPort int  `ini:"proxy_port"`
	isEnable  bool `ini:"-"`
}

func (b *PluginHttpProxy) OnInit() {
	b.isEnable = false
	if b.ProxyPort < 0 || b.ProxyPort > 65535 {
		panic("invalid local port[0-65535]")
	}
	b.isEnable = true
}
