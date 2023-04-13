package cmd

import (
	"path/filepath"

	"github.com/ferama/crauti/pkg/admin"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/ferama/crauti/pkg/gateway/kube"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

var log *zerolog.Logger

func init() {
	// this one is here to make some init vars available to other
	// init functions.
	// The use case is the CRAUTI_DEBUG that need to be available as
	// soon as possibile in order to instantiate the logger correctly
	viper.ReadInConfig()
	conf.Update()

	log = logger.GetLogger("root")

	if home := homedir.HomeDir(); home != "" {
		rootCmd.Flags().StringP("kubeconfig", "k", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.Flags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	}

	rootCmd.Flags().StringP("config", "c", "", "set config file path")
}

var rootCmd = &cobra.Command{
	Use: "crauti",
	Run: func(cmd *cobra.Command, args []string) {
		// the api gateway server
		// log.Info().Msgf("gateway listening on '%s'", conf.ConfInst.Gateway.ListenAddress)
		gwServer := gateway.NewGateway(":80", ":443")

		var stopper chan struct{}
		update := func(e fsnotify.Event) {
			if stopper != nil {
				close(stopper)
			}
			// WARN: Reset is marked as intended for tests only
			//
			viper.Reset()
			config, _ := cmd.Flags().GetString("config")
			if config != "" {
				viper.SetConfigFile(config)
			}
			conf.Reset()

			stopper = make(chan struct{})

			log.Print("config file changed: ", e.Name)
			err := viper.ReadInConfig()
			if err != nil {
				log.Print("cannot read config file")
			}
			conf.Update()
			gwServer.Update()

			if conf.ConfInst.Gateway.Kubernetes.Autodiscover {
				kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
				// the kubernetes services informer
				kube.NewObserver(gwServer, kubeconfig, stopper)
				log.Info().Msgf("k8s service informer started")
			}
		}

		update(fsnotify.Event{Name: ""})

		viper.OnConfigChange(update)
		viper.WatchConfig()

		// setupAdminServer(gwServer)
		adminServer := admin.NewAdminServer()
		go adminServer.Start()

		// start the gateway server
		gwServer.Start()

		if stopper != nil {
			close(stopper)
		}
	},
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}
