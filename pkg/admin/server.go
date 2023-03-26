package admin

import (
	"github.com/ferama/crauti/pkg/admin/api"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

type adminServer struct {
	router *gin.Engine
}

func NewAdminServer() *adminServer {

	// Install admin apis
	gin.SetMode(gin.ReleaseMode)
	ginrouter := gin.New()
	ginrouter.Use(
		// do not log k8s calls to health
		// gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)

	s := &adminServer{
		router: ginrouter,
	}
	s.setupRoutes()

	log.Info().Msgf("admin listening on '%s'", conf.ConfInst.AdminApiListenAddress)
	return s
}

func (s *adminServer) setupRoutes() {
	// setup health endpoint
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})
	// install the prometheus metrics endpoint
	s.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// install the /api routes
	api.RootRouter(s.router.Group("/api"))
}

func (s *adminServer) Start() {
	s.router.Run(conf.ConfInst.AdminApiListenAddress)
}
