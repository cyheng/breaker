package feature

const (
	FPortal = "portal"
)

func (c *PortalConfig) OnInit() {
	if c.ServerAddr == "" {
		c.ServerAddr = "0.0.0.0:80"
	}
}
func (b *PortalConfig) ServiceName() string {
	return FPortal
}
func init() {
	RegisterConfig(FPortal, &PortalConfig{})
}
