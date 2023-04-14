package api

import (
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
)

type adminGroup struct{}

// Routes setup the root api routes
func adminRoutes(router *gin.RouterGroup) {
	r := &adminGroup{}

	router.GET("config", r.config)
}

func (r *adminGroup) config(c *gin.Context) {
	c.Negotiate(http.StatusOK, gin.Negotiate{
		Data:    conf.ConfInst,
		Offered: supportedFormats,
	})
}
