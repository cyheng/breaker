package feature

//Portal implement the feature interface
type Portal struct {
	ServerAddr string
}

type PortalConfig struct {
	ServerAddr string `ini:"server_addr"`
}

func (c *PortalConfig) OnInit() {
	if c.ServerAddr == "" {
		c.ServerAddr = "0.0.0.0:80"
	}
}

func (c *PortalConfig) NewFeature() (Feature, error) {
	res := &Portal{}
	res.ServerAddr = c.ServerAddr
	return res, nil
}

func init() {
	RegisterConfig("portal", &PortalConfig{})

}

func (p *Portal) Name() string {
	return "portal"
}
