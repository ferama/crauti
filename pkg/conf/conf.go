package conf

import (
	"log"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ConfInst config

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

// middelewares configuration struct
type middlewares struct {
	Cors  cors  `yaml:"cors"`
	Cache Cache `yaml:"cache"`
	Proxy proxy `yaml:"proxy"`
	// on timeout expiration, the context will be canceled and request
	// aborted. Use -1 or any value lesser than 0 to disable timeout
	Timeout time.Duration `yaml:"timeout,omitempty"`
}

func (m *middlewares) clone() middlewares {
	c := middlewares{
		Cors:    m.Cors.clone(),
		Cache:   m.Cache.clone(),
		Proxy:   m.Proxy.clone(),
		Timeout: m.Timeout,
	}
	return c
}

type kubernetes struct {
	// if service discover is enabled or not
	Autodiscover bool `yaml:"autodiscover"`
	// if not empyt, limits discovery to the specified namespace
	WatchNamespace string `yaml:"watchNamespace"`
}

type gateway struct {
	ListenAddress string        `yaml:"listenAddress"`
	WriteTimeout  time.Duration `yaml:"writeTimeout"`
	ReadTimeout   time.Duration `yaml:"readTimeout"`
	IdleTimeout   time.Duration `yaml:"idleTimeout"`
}

// config holds all the config values
type config struct {
	// debug log level
	Debug bool `yaml:"debug"`
	// Listeners conf
	Gateway               gateway `yaml:"gateway"`
	AdminApiListenAddress string  `yaml:"adminApiListenAddress"`
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

func setDefaults() {
	viper.SetDefault("Debug", false)

	// Gateway conf
	viper.SetDefault("Gateway.ListenAddress", ":8080")
	viper.SetDefault("Gateway.ReadTimeout", "120s")
	viper.SetDefault("Gateway.WriteTimeout", "120s")
	viper.SetDefault("Gateway.IdleTimeout", "360s")

	// viper.SetDefault("Kubernetes.Autodiscover", true)
	viper.SetDefault("AdminApiListenAddress", ":8181")

	///////////////////////////////////////////////////////
	//
	// Middlewares defaults
	//
	///////////////////////////////////////////////////////

	// Cors defaults
	viper.SetDefault("Middlewares.Cors.Enabled", false)

	// Timeout defaults
	// this timeout acts like the Gateway.WriteTimeout but it can be set
	// per mountPoint
	viper.SetDefault("Middlewares.Timeout", "-1") // disabled by default

	// Reverse Proxy defualts
	viper.SetDefault("Middlewares.Proxy.PreserveHostHeader", true)

	// Cache defaults
	viper.SetDefault("Middlewares.Cache.Enabled", false)
	viper.SetDefault("Middlewares.Cache.Redis.Host", "localhost")
	viper.SetDefault("Middlewares.Cache.Redis.Port", 6379)
	viper.SetDefault("Middlewares.Cache.TTL", "5m")
	viper.SetDefault("Middlewares.Cache.Methods", "GET,HEAD,OPTIONS")
}

func init() {
	viper.SetConfigName("crauti")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// this two lines enables set config through env vars.
	// you can use something like
	//	CRAUTI_YOURCONFVARHERE=YOURVALUE
	viper.AutomaticEnv()
	viper.SetEnvPrefix("crauti")

	setDefaults()
}

func Update() {
	ConfInst.reset()

	err := viper.Unmarshal(&ConfInst)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	// merge mountpoints middleware configuration
	for idx, i := range ConfInst.MountPoints {
		// m is the default(global) Middlewares configuration
		m := ConfInst.Middlewares.clone()

		b, _ := yaml.Marshal(i.Middlewares)
		// I'm using unmarshal here to merge the mountPoint specific configuration
		// into the global one and then some row above, to assign the merged conf
		// to the monutPoint. There are some quirks that needs to be managed
		//
		// 	1. slices needs manually merging logic (Cache.merge methods for example)
		//	2. boolean needs to be treated like pointer to boolean to reflect the three
		// 	   available states: true, false, nil/undefined
		yaml.Unmarshal(b, &m)

		m.Cache.merge(i.Middlewares.Cache)

		ConfInst.MountPoints[idx].Middlewares = m
	}

	if ConfInst.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// debug utility
func Dump() (string, error) {
	b, err := yaml.Marshal(ConfInst)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
