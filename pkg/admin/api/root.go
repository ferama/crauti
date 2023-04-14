package api

import (
	"github.com/gin-gonic/gin"
)

var supportedFormats = []string{gin.MIMEJSON, gin.MIMEYAML}

func RootRouter(router *gin.RouterGroup) {
	configRoutes(router.Group("/config"))
	cacheRoutes(router.Group("/cache"))
	mountPointRoutes(router.Group("/mount-point"))
}
