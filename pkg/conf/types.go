package conf

type MountPoint struct {
	// crauti gateway mount path
	// like /api/config
	Path string `yaml:"path"`
	// full upstream definition
	// like http://my-service.my-namespace:port
	Upstream string `yaml:"upstream"`
}

type Kubernetes struct {
	Autodiscover   bool   `yaml:"autodiscover"`
	WatchNamespace string `yaml:"watchNamespace"`
}

// config holds all the config values
type config struct {
	MountPoints []MountPoint `yaml:"mountPoints"`
	// if the service informer is enabled or not
	Kubernetes Kubernetes `yaml:"kubernetes"`

	GatewayListenAddress  string `yaml:"gatewayListenAddress"`
	AdminApiListenAddress string `yaml:"adminApiListenAddress"`
}

// resets the config fields. called on dynamic conf update
func (c *config) reset() {
	c.MountPoints = nil
}
