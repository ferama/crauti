package conf

import (
	"log"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var ConfInst config

func setDefaults() {
	viper.SetDefault("K8sAutodiscover", true)
	viper.SetDefault("GatewayListenAddress", ":8080")
	viper.SetDefault("AdminApiListenAddress", ":9090")

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
		m := ConfInst.Middlewares
		b, _ := yaml.Marshal(i.Middlewares)
		yaml.Unmarshal(b, &m)

		m.Cache.merge(i.Middlewares.Cache)

		ConfInst.MountPoints[idx].Middlewares = m
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
