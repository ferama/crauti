package main

import (
	"log"
	"time"

	"github.com/ferama/crauti/pkg/admin"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/ferama/crauti/pkg/kube"
	"github.com/gin-gonic/gin"
)

func main() {
	go func() {
		for {
			conf.Dump()
			time.Sleep(3 * time.Second)
		}
	}()

	// the api gateway server
	log.Printf("gateway listening on '%s'", conf.Config.GatewayListenAddress)
	gwServer := gateway.NewServer(conf.Config.GatewayListenAddress)

	if conf.Config.K8sAutodiscover {
		// stop signal for the informer
		stopper := make(chan struct{})
		defer close(stopper)
		// the kubernetes services informer
		kube.NewSvcHandler(gwServer, stopper)
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
}
