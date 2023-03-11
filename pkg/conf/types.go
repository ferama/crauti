package conf

type MountPoint struct {
	Path     string
	Upstream string
}

// Config holds all the config values
type Config struct {
	MountPoints     []MountPoint
	K8sAutodiscover bool
}
