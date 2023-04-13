package api

import (
	"net/http"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type mountPointGroup struct{}

func mountPointRoutes(router *gin.RouterGroup) {
	r := mountPointGroup{}
	router.GET("", r.get)
	router.POST("", r.post)
	router.PUT("", r.put)
	router.DELETE("", r.delete)
}
func (r *mountPointGroup) filter(path, host string) []conf.MountPoint {
	res := make([]conf.MountPoint, 0)
	for _, m := range conf.ConfInst.MountPoints {
		if m.Path == path && m.MatchHost == host {
			res = append(res, m)
		}
	}
	return res
}

func (r *mountPointGroup) get(c *gin.Context) {
	path := c.Query("path")
	host := c.Query("host")

	if path == "" && host == "" {
		c.YAML(http.StatusOK, conf.ConfInst.MountPoints)
		return
	}

	c.YAML(http.StatusOK, r.filter(path, host))
}

func (r *mountPointGroup) post(c *gin.Context) {
}

func (r *mountPointGroup) put(c *gin.Context) {
}

func (r *mountPointGroup) delete(c *gin.Context) {
	path := c.Query("path")
	host := c.Query("host")
	if path != "" {
		res := make([]conf.MountPoint, 0)
		for _, m := range conf.ConfInst.MountPoints {
			if m.Path != path || m.MatchHost != host {
				res = append(res, m)
			}
		}
		viper.Set("MountPoints", res)
		viper.WriteConfig()
	}
}
