package main

import (
	"log"

	"github.com/ferama/crauti/pkg/admin"
	"github.com/ferama/crauti/pkg/gateway/server"
	"github.com/ferama/crauti/pkg/kube"
	"github.com/gin-gonic/gin"
)

func main() {
	// the api gateway server
	gwServer := server.NewServer(":8080")

	// stop signal for the informer
	stopper := make(chan struct{})
	defer close(stopper)
	// the kubernetes services informer
	kube.NewSvcHandler(gwServer, stopper)

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

	adminApiListenAddr := ":9000"
	log.Printf("Admin api listening on '%s'", adminApiListenAddr)
	go ginrouter.Run(adminApiListenAddr)

	// start the gateway server
	gwServer.Start()
}
