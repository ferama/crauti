package api

import (
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
)

type configGroup struct{}

// Routes setup the root api routes
func configRoutes(router *gin.RouterGroup) {
	r := &configGroup{}

	router.GET("", r.config)
}

func (r *configGroup) config(c *gin.Context) {
	c.Negotiate(http.StatusOK, gin.Negotiate{
		Data:    conf.ConfInst,
		Offered: supportedFormats,
	})
}
