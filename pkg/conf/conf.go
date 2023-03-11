package conf

import (
	"log"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var Config config

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
	Config.reset()

	err := viper.Unmarshal(&Config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func Dump() (string, error) {
	// b, err := json.MarshalIndent(Crauti, "", "    ")
	b, err := yaml.Marshal(Config)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
