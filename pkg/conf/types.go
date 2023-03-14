package conf

import "time"

type MountPoint struct {
	// crauti gateway mount path
	// like /api/config
	Path string `yaml:"path"`
	// full upstream definition
	// like http://my-service.my-namespace:port
	Upstream string `yaml:"upstream"`
	// middlewares configuration can be overridden setting
	// changed values here
	Middlewares middlewares `yaml:"middlewares"`
}

type kubernetes struct {
	// if service discover is enabled or not
	Autodiscover bool `yaml:"autodiscover"`
	// if not empyt, limits discovery to the specified namespace
	WatchNamespace string `yaml:"watchNamespace"`
}

// /////////////////////////////////////////////////////////////////////////////
//
//	Start Middlewares configuration
//
// /////////////////////////////////////////////////////////////////////////////

type cors struct {
	Enabled bool `yaml:"enabled"`

	// TODO: remove this. for test only
	// omitempty here is important, merge will fail otherwise
	//
	Val string `yaml:"val,omitempty"`
}

type redis struct {
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type cache struct {
	Enabled    bool          `yaml:"enabled"`
	Redis      redis         `yaml:"redis,omitempty"`
	TTL        time.Duration `yaml:"cacheTTL,omitempty"`
	Methods    []string      `yaml:"methods,omitempty"`
	KeyHeaders []string      `yaml:"keyHeaders,omitempty"`
}

// slice types needs manually merging logic
// When not defined (nil case) we should use the global values
// If defined but empty ([] case), we should use a nil value
func (c *cache) merge(target cache) {
	if target.Methods == nil {
		c.Methods = ConfInst.Middlewares.Cache.Methods
	} else if len(target.Methods) == 0 {
		c.Methods = nil
	}

	if target.KeyHeaders == nil {
		c.KeyHeaders = ConfInst.Middlewares.Cache.KeyHeaders
	} else if len(target.KeyHeaders) == 0 {
		c.KeyHeaders = nil
	}
}

// middelewares configuration struct
type middlewares struct {
	Cors  cors  `yaml:"cors"`
	Cache cache `yaml:"cache"`
}

// /////////////////////////////////////////////////////////////////////////////
//
//	End Middlewares configuration
//
// /////////////////////////////////////////////////////////////////////////////

// config holds all the config values
type config struct {
	// Listeners conf
	GatewayListenAddress  string `yaml:"gatewayListenAddress"`
	AdminApiListenAddress string `yaml:"adminApiListenAddress"`
	// kubernetes relatech conf
	Kubernetes kubernetes `yaml:"kubernetes"`
	// global middlewares configuration
	Middlewares middlewares `yaml:"middlewares"`
	// define mount points
	MountPoints []MountPoint `yaml:"mountPoints"`
}

// resets the config fields. called on dynamic conf update
func (c *config) reset() {
	c.MountPoints = nil
}
