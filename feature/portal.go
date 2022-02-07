package feature

type PortalConfig struct {
	LoggerConfig `ini:"Logger"`
	ServerAddr   string `ini:"server_addr"`
}

func (c *PortalConfig) OnInit() {
	c.LoggerConfig.OnInit()
	if c.ServerAddr == "" {
		c.ServerAddr = "0.0.0.0:80"
	}
}
