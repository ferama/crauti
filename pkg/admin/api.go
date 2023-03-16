package admin

import (
	"net/http"
	"strings"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/ferama/crauti/pkg/gateway"
	"github.com/gin-gonic/gin"
)

type adminRoutes struct {
	gwServer *gateway.Server
}

// Routes setup the root api routes
func Routes(gwServer *gateway.Server, router *gin.RouterGroup) {
	r := &adminRoutes{
		gwServer: gwServer,
	}

	router.GET("health", r.Health)
	router.GET("config/:encoding", r.Config)
}

func (r *adminRoutes) Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func (r *adminRoutes) Config(c *gin.Context) {
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
