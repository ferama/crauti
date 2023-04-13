package api

import (
	"fmt"
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
)

type mountPointGroup struct{}

func mountPointRoutes(router *gin.RouterGroup) {
	r := mountPointGroup{}
	router.GET("", r.get)
}
func (r *mountPointGroup) filter(path, host string) []conf.MountPoint {
	res := make([]conf.MountPoint, 0)
	// for _, m := range conf.ConfInst.MountPoints {
	// 	// if m.Path == path && m
	// }
	return res
}

func (r *mountPointGroup) get(c *gin.Context) {
	path := c.Query("path")
	host := c.Query("host")

	if path == "" && host == "" {
		c.JSON(http.StatusOK, conf.ConfInst.MountPoints)
		return
	}

	message := fmt.Sprintf("path: %s, host: %s", path, host)
	c.String(http.StatusOK, message)

}
