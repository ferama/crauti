package cmd

import (
	"path/filepath"

	"github.com/ferama/crauti/pkg/admin"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/ferama/crauti/pkg/kube"
	"github.com/ferama/crauti/pkg/logger"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

var log *zerolog.Logger

func init() {
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
		config, _ := cmd.Flags().GetString("config")
		if config != "" {
			viper.SetConfigFile(config)
		}
		err := viper.ReadInConfig() // Find and read the config file
		if err != nil {             // Handle errors reading the config file
			log.Print("no config file detected, using default values")
		}
		conf.Update()

		// the api gateway server
		log.Info().Msgf("gateway listening on '%s'", conf.ConfInst.Gateway.ListenAddress)
		gwServer := gateway.NewServer(conf.ConfInst.Gateway.ListenAddress)

		if conf.ConfInst.Kubernetes.Autodiscover {
			kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
			// stop signal for the informer
			stopper := make(chan struct{})
			defer close(stopper)
			// the kubernetes services informer
			kube.NewSvcHandler(gwServer, kubeconfig, stopper)
			log.Info().Msgf("k8s service informer started")
		} else {
			gwServer.UpdateHandlers()
			viper.OnConfigChange(func(e fsnotify.Event) {
				log.Print("config file changed:", e.Name)
				conf.Update()
				gwServer.UpdateHandlers()
			})
			viper.WatchConfig()
		}

		// setupAdminServer(gwServer)
		adminServer := admin.NewAdminServer()
		go adminServer.Start()

		// start the gateway server
		gwServer.Start()
	},
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}
