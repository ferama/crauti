package admin

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/ferama/crauti/pkg/admin/api"
	"github.com/ferama/crauti/pkg/admin/ui"
	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-contrib/cors"
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
	ginrouter.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Content-Type, Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	s := &adminServer{
		router: ginrouter,
	}
	s.setupRoutes()

	// static files custom middleware
	// use the "dist" dir (the vite target) as static root
	fsRoot, _ := fs.Sub(ui.StaticFiles, "dist")
	fileserver := http.FileServer(http.FS(fsRoot))
	ginrouter.Use(func(c *gin.Context) {
		fileserver.ServeHTTP(c.Writer, c.Request)
		c.Abort()
	})

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
