package conf

import (
	"log"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var Crauti config

func setDefaults() {
	viper.SetDefault("K8sAutodiscover", true)
	viper.SetDefault("GatewayListenAddress", ":8080")
	viper.SetDefault("AdminApiListenAddress", ":9090")
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
	Crauti.reset()

	err := viper.Unmarshal(&Crauti)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	// merge mountpoints middleware configuration
	for idx, i := range Crauti.MountPoints {
		m := Crauti.Middlewares
		b, _ := yaml.Marshal(i.Middlewares)
		yaml.Unmarshal(b, &m)
		Crauti.MountPoints[idx].Middlewares = m
	}
}

// debug utility
func Dump() (string, error) {
	b, err := yaml.Marshal(Crauti)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
