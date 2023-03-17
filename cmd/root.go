package cmd

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/ferama/crauti/pkg/admin"
	"github.com/ferama/crauti/pkg/cache"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/ferama/crauti/pkg/kube"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/util/homedir"
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		rootCmd.Flags().StringP("kubeconfig", "k", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.Flags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	}
	rootCmd.Flags().StringP("config", "c", "", "set config file path")

	rootCmd.Flags().BoolP("debug", "d", false, "debug")
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
			log.Println("no config file detected, using default values")
		}
		conf.Update()

		debug, _ := cmd.Flags().GetBool("debug")
		if debug {
			go func() {
				for {
					c, _ := conf.Dump()
					log.Printf("current conf:\n\n%s\n", c)
					time.Sleep(3 * time.Second)
				}
			}()
		}

		// the api gateway server
		log.Printf("gateway listening on '%s'", conf.ConfInst.GatewayListenAddress)
		gwServer := gateway.NewServer(conf.ConfInst.GatewayListenAddress)

		if conf.ConfInst.Kubernetes.Autodiscover {
			kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
			// stop signal for the informer
			stopper := make(chan struct{})
			defer close(stopper)
			// the kubernetes services informer
			kube.NewSvcHandler(gwServer, kubeconfig, stopper)
			log.Println("k8s service informer started")
		} else {
			gwServer.UpdateHandlers()
			viper.OnConfigChange(func(e fsnotify.Event) {
				fmt.Println("config file changed:", e.Name)
				conf.Update()
				gwServer.UpdateHandlers()
			})
			viper.WatchConfig()
		}

		// Install admin apis
		gin.SetMode(gin.ReleaseMode)
		ginrouter := gin.New()
		ginrouter.Use(
			// do not log k8s calls to health
			// gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
			gin.Recovery(),
		)
		// we could also mount the gin router into the default mux
		// but a dedicated port could be a better choice. The idea here
		// is to leave this api/port not exposed directly.
		admin.Routes(gwServer, ginrouter.Group("/"))
		cache.Routes(ginrouter.Group("/cache"))

		log.Printf("admin api listening on '%s'", conf.ConfInst.AdminApiListenAddress)
		go ginrouter.Run(conf.ConfInst.AdminApiListenAddress)

		// start the gateway server
		gwServer.Start()
	},
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}
