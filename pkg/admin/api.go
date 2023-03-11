package admin

import (
	"net/http"

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
	router.GET("routes", r.Routes)
}

func (r *adminRoutes) Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "ok",
	})
}

func (r *adminRoutes) Routes(c *gin.Context) {
	// mountPoints := r.gwServer.GetMountpoints()
	type responseItem struct {
		Upstream string `json:"Upstream"`
		Path     string `json:"Path"`
	}
	var response []responseItem

	for _, route := range conf.Crauti.MountPoints {
		response = append(response, responseItem{
			route.Upstream,
			route.Path,
		})
	}

	c.JSON(http.StatusOK, response)
}
