package api

import (
	"github.com/gin-gonic/gin"
)

func RootRouter(router *gin.RouterGroup) {
	adminRoutes(router.Group("/"))
	cacheRoutes(router.Group("/cache"))
}
