package api

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RootRouter(router *gin.Engine) {

	// setup health endpoint
	router.GET("health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "ok",
		})
	})

	// install the prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	apiRouter := router.Group("/api")

	adminRoutes(apiRouter.Group("/"))
	cacheRoutes(apiRouter.Group("/cache"))
}
