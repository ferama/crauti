package conf

import (
	"strings"
	"time"

	"github.com/ferama/crauti/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	DefaultMaxRequestBodySize string = "10mb"
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
	Middlewares Middlewares `yaml:"middlewares"`
}

// middelewares configuration struct
type Middlewares struct {
	Cors  Cors  `yaml:"cors"`
	Cache Cache `yaml:"cache"`
	// on timeout expiration, the context will be canceled and request
	// aborted. Use -1 or any value lesser than 0 to disable timeout
	Timeout time.Duration `yaml:"timeout,omitempty"`
	// Use -1 or any value lesser than 0 to disable the limit
	MaxRequestBodySize string `yaml:"maxRequestBodySize,omitempty"`
	// default false. If true, the ReverseProxy will not set
	// the Host header to the real upstream host while forwarding the
	// request to the upstream
	PreserveHostHeader *bool `yaml:"preserveHostHeader,omitempty"`
	// VirtualHost like behaviour
	MatchHost string `yaml:"matchHost,omitempty"`
	// if true, all requeste will be redirected to https
	RedirectToHTTPS *bool `yaml:"redirectToHTTPS,omitempty"`
	// set rewrite parameters
	Rewrite rewrite `yaml:"rewrite,omitempty"`
	// if not empty, enables the jwt auth middleware
	JwksURL string `yaml:"jwksURL,omitempty"`
}

// Helper function that check for nil value on Enabled field
func (m *Middlewares) IsPreserveHostHeader() bool {
	return m.PreserveHostHeader != nil && *m.PreserveHostHeader
}

// Helper function that check for nil value on Enabled field
func (m *Middlewares) IsRedirectToHTTPS() bool {
	return m.RedirectToHTTPS != nil && *m.RedirectToHTTPS
}

func (m *Middlewares) clone() Middlewares {
	preserveHostHeader := *m.PreserveHostHeader
	redirectToHTTPS := *m.RedirectToHTTPS

	c := Middlewares{
		Cors:               m.Cors.clone(),
		Cache:              m.Cache.clone(),
		Timeout:            m.Timeout,
		MaxRequestBodySize: m.MaxRequestBodySize,
		PreserveHostHeader: &preserveHostHeader,
		RedirectToHTTPS:    &redirectToHTTPS,
		MatchHost:          m.MatchHost,
		Rewrite:            m.Rewrite.clone(),
		JwksURL:            m.JwksURL,
	}
	return c
}

type kubernetes struct {
	// if service discover is enabled or not
	Autodiscover bool `yaml:"autodiscover"`
	// if not empyt, limits discovery to the specified namespace
	WatchNamespace string `yaml:"watchNamespace"`
}

type keyPair struct {
	FullChain  string `yaml:"fullChain"`
	PrivateKey string `yaml:"key"`
}
type gateway struct {
	WriteTimeout time.Duration `yaml:"writeTimeout"`
	ReadTimeout  time.Duration `yaml:"readTimeout"`
	IdleTimeout  time.Duration `yaml:"idleTimeout"`
	// if true enable the HTTPS server. You need to provide
	// at least a keypair
	HTTPSEnabled bool `yaml:"httpsEnabled"`
	// keyPairs paths to load certificates from if autoHTTPS is disabled
	KeyPairs []keyPair `yaml:"keyPairs"`
	// if true enable autoCert. Implies HTTPSEnabled = true
	AutoHTTPSEnabled bool `yaml:"autoHTTPSEnabled"`
	// by default AutoHTTPS uses the kubernetes secret backend to store
	// certificate cache. If you enable this flag, a local dir will be
	// used instead
	AutoHTTPSUseLocalDir bool `yaml:"autoHTTPSUseLocalDir"`
	// set a local dir to use to cache certificates
	// default to: ./certs-cache
	AutoHTTPSLocalDir string `yaml:"autoHTTPSLocalDir"`
	// kubernetes related conf
	Kubernetes kubernetes `yaml:"kubernetes"`
}

// config holds all the config values
type config struct {
	// debug log level
	Debug bool `yaml:"debug"`
	// Listeners conf
	Gateway               gateway `yaml:"gateway"`
	AdminApiListenAddress string  `yaml:"adminApiListenAddress"`
	// global middlewares configuration
	Middlewares Middlewares `yaml:"middlewares"`
	// define mount points
	MountPoints []MountPoint `yaml:"mountPoints"`
}

// resets the config fields. called on dynamic conf update
func (c *config) reset() {
	c.MountPoints = nil
}

func setDefaults() {
	// IMPORTANT:
	// conf that doesn't have a default actually can't be set with
	// env vars.
	// Example:
	//		CRAUTI_KUBERNETES_WATCHNAMESPACE="test"
	// will work, because it has a default here

	viper.SetDefault("Debug", false)

	// Gateway conf
	viper.SetDefault("Gateway.ListenAddress", ":8080")
	viper.SetDefault("Gateway.ReadTimeout", "120s")
	viper.SetDefault("Gateway.WriteTimeout", "120s")
	viper.SetDefault("Gateway.IdleTimeout", "360s")
	viper.SetDefault("Gateway.HTTPSEnabled", false)
	viper.SetDefault("Gateway.AutoHTTPSEnabled", false)
	viper.SetDefault("Gateway.AutoHTTPSUseLocalDir", false)
	viper.SetDefault("Gateway.AutoHTTPSLocalDir", "./certs-cache")
	viper.SetDefault("Gateway.Kubernetes.Autodiscover", false)
	viper.SetDefault("Gateway.Kubernetes.WatchNamespace", "")

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
	viper.SetDefault("Middlewares.Timeout", "-1s") // disabled by default
	viper.SetDefault("Middlewares.MaxRequestBodySize", DefaultMaxRequestBodySize)
	viper.SetDefault("Middlewares.PreserveHostHeader", true)
	viper.SetDefault("Middlewares.RedirectToHTTPS", false)
	viper.SetDefault("Middlewares.MatchHost", "") // disabled by default
	viper.SetDefault("Middlewares.JwksURL", "")   // disabled by default

	// Cache defaults
	viper.SetDefault("Middlewares.Cache.Enabled", false)
	viper.SetDefault("Middlewares.Cache.Redis.Host", "localhost")
	viper.SetDefault("Middlewares.Cache.Redis.Port", 6379)
	viper.SetDefault("Middlewares.Cache.TTL", "5m")
	viper.SetDefault("Middlewares.Cache.Methods", "GET,HEAD,OPTIONS")
	viper.SetDefault("Middlewares.Cache.KeyHeaders", "")
}

func init() {
	viper.SetConfigName("crauti")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
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
		log.Error().Msgf("unable to decode into struct, %v", err)
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

		_, err = utils.ConvertToBytes(m.MaxRequestBodySize)
		if err != nil {
			m.MaxRequestBodySize = DefaultMaxRequestBodySize
			log.Error().Msgf("unable to parse MaxRequestBodySize. mountPath: '%s'. reverting to default", i.Path)
		}

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
