package conf

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

var CrautiConf Config

func init() {
	viper.SetConfigName("crauti")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// this two lines enables set config through env vars.
	// you can use something like
	//	CRAUTI_YOURCONFVARHERE=YOURVALUE
	viper.AutomaticEnv()
	viper.SetEnvPrefix("crauti")

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		fmt.Println(fmt.Errorf("fatal error config file: %w", err))
		os.Exit(1)
	}

	err = viper.Unmarshal(&CrautiConf)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}
}
