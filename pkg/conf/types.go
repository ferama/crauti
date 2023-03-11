package conf

type MountPoint struct {
	Path     string
	Upstream string
}

// config holds all the config values
type config struct {
	MountPoints           []MountPoint
	K8sAutodiscover       bool
	GatewayListenAddress  string
	AdminApiListenAddress string
}
