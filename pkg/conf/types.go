package conf

type MountPoint struct {
	// crauti gateway mount path
	// like /api/config
	Path string `yaml:"path"`
	// full upstream definition
	// like http://my-service.my-namespace:port
	Upstream string `yaml:"upstream"`
}

// config holds all the config values
type config struct {
	MountPoints []MountPoint `yaml:"mountPoints"`
	// if the service informer is enabled or not
	K8sAutodiscover bool `yaml:"k8sAutodiscover"`

	GatewayListenAddress  string `yaml:"gatewayListenAddress"`
	AdminApiListenAddress string `yaml:"adminApiListenAddress"`
}

func (c *config) reset() {
	c.MountPoints = nil
}
