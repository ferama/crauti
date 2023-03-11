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

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Println("no config file detected, using default values")
	}
	Update()
}

func Update() {
	err := viper.Unmarshal(&Crauti)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}

func Dump() error {
	// b, err := json.MarshalIndent(Crauti, "", "    ")
	b, err := yaml.Marshal(Crauti)
	if err != nil {
		return err
	}
	log.Printf("Current conf:\n\n%s\n", string(b))
	return nil
}
