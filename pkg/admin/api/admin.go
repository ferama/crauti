package api

import (
	"net/http"
	"strings"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
)

type adminGroup struct{}

// Routes setup the root api routes
func adminRoutes(router *gin.RouterGroup) {
	r := &adminGroup{}

	router.GET("config", r.config)
	router.GET("config/:encoding", r.config)
}

func (r *adminGroup) config(c *gin.Context) {
	type binding struct {
		Encoding string `uri:"encoding"`
	}
	var encoding binding
	if err := c.ShouldBindUri(&encoding); err == nil {
		if strings.ToLower(encoding.Encoding) == "yaml" {
			c.YAML(http.StatusOK, conf.ConfInst)
			return
		}
	}
	c.JSON(http.StatusOK, conf.ConfInst)
}
