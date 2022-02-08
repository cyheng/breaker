package feature

type PortalConfig struct {
	LoggerConfig `ini:"Logger"`
	// 不带分组
	PluginHttpProxy `ini:"DEFAULT,omitempty"`
	ServerAddr      string `ini:"server_addr"`
}

func (c *PortalConfig) OnInit() {
	c.LoggerConfig.OnInit()
	c.PluginHttpProxy.OnInit()
	if c.ServerAddr == "" {
		c.ServerAddr = "0.0.0.0:80"
	}
}
