package cmd

import (
	"log"
	"path/filepath"
	"time"

	"github.com/ferama/crauti/pkg/admin"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/ferama/crauti/pkg/kube"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

func init() {
	if home := homedir.HomeDir(); home != "" {
		rootCmd.Flags().StringP("kubeconfig", "k", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		rootCmd.Flags().StringP("kubeconfig", "k", "", "absolute path to the kubeconfig file")
	}
}

var rootCmd = &cobra.Command{
	Use: "crauti",
	Run: func(cmd *cobra.Command, args []string) {
		go func() {
			for {
				c, _ := conf.Dump()
				log.Printf("current conf:\n\n%s\n", c)
				time.Sleep(3 * time.Second)
			}
		}()

		// the api gateway server
		log.Printf("gateway listening on '%s'", conf.Config.GatewayListenAddress)
		gwServer := gateway.NewServer(conf.Config.GatewayListenAddress)

		if conf.Config.K8sAutodiscover {
			kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
			// stop signal for the informer
			stopper := make(chan struct{})
			defer close(stopper)
			// the kubernetes services informer
			kube.NewSvcHandler(gwServer, kubeconfig, stopper)
			log.Println("k8s service informer started")
		}

		// Install admin apis
		gin.SetMode(gin.ReleaseMode)
		ginrouter := gin.New()
		ginrouter.Use(
			// do not log k8s calls to health
			gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
			gin.Recovery(),
		)
		// we could also mount the gin router into the default mux
		// but a dedicated port could be a better choice. The idea here
		// is to leave this api/port not exposed directly.
		admin.Routes(gwServer, ginrouter.Group("/"))

		log.Printf("admin api listening on '%s'", conf.Config.AdminApiListenAddress)
		go ginrouter.Run(conf.Config.AdminApiListenAddress)

		// start the gateway server
		gwServer.Start()
	},
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}
