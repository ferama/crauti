package api

import (
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type mountPointGroup struct {
	mu sync.Mutex
}

func mountPointRoutes(router *gin.RouterGroup) {
	r := mountPointGroup{}
	router.GET("", r.get)
	router.POST("", r.post)
	router.PUT("", r.put)
	router.DELETE("", r.delete)
}

func (r *mountPointGroup) search(path, host string) []conf.MountPoint {
	res := make([]conf.MountPoint, 0)
	for _, m := range conf.ConfInst.MountPoints {
		if m.Path == path && m.MatchHost == host {
			res = append(res, m)
		}
	}
	return res
}

func (r *mountPointGroup) get(c *gin.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	path := c.Query("path")
	host := c.Query("host")

	if path == "" && host == "" {
		c.Negotiate(http.StatusOK, gin.Negotiate{
			Data:    conf.ConfInst.MountPoints,
			Offered: supportedFormats,
		})
		return
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		Data:    r.search(path, host),
		Offered: supportedFormats,
	})
}

func (r *mountPointGroup) post(c *gin.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var mp conf.MountPoint
	err := c.BindJSON(&mp)
	if err != nil {
		c.YAML(http.StatusInternalServerError, err)
	}
	result := r.search(mp.Path, mp.MatchHost)

	if len(result) > 0 {
		c.Negotiate(http.StatusBadRequest, gin.Negotiate{
			Data:    "already exists",
			Offered: supportedFormats,
		})
		return
	}

	mountPoints := make([]conf.MountPoint, 0)
	mountPoints = append(mountPoints, conf.ConfInst.MountPoints...)
	mountPoints = append(mountPoints, mp)

	viper.Set("MountPoints", mountPoints)
	err = viper.WriteConfig()
	if err != nil {
		c.Negotiate(http.StatusInternalServerError, gin.Negotiate{
			Data:    "can't update config",
			Offered: supportedFormats,
		})
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		Data:    mp,
		Offered: supportedFormats,
	})
}

func (r *mountPointGroup) put(c *gin.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var mp conf.MountPoint
	err := c.BindJSON(&mp)
	if err != nil {
		c.YAML(http.StatusInternalServerError, err)
	}
	result := r.search(mp.Path, mp.MatchHost)

	if len(result) == 0 {
		c.Negotiate(http.StatusBadRequest, gin.Negotiate{
			Data:    "mount point doesn't exists",
			Offered: supportedFormats,
		})
		return
	}

	mountPoints := make([]conf.MountPoint, 0)
	for _, m := range conf.ConfInst.MountPoints {
		if m.Path != result[0].Path || m.MatchHost != result[0].MatchHost {
			mountPoints = append(mountPoints, m)
		}
	}
	mountPoints = append(mountPoints, mp)

	viper.Set("MountPoints", mountPoints)
	err = viper.WriteConfig()
	if err != nil {
		c.Negotiate(http.StatusInternalServerError, gin.Negotiate{
			Data:    "can't update config",
			Offered: supportedFormats,
		})
	}

	c.Negotiate(http.StatusOK, gin.Negotiate{
		Data:    mp,
		Offered: supportedFormats,
	})
}

func (r *mountPointGroup) delete(c *gin.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

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
		err := viper.WriteConfig()
		if err != nil {
			c.Negotiate(http.StatusInternalServerError, gin.Negotiate{
				Data:    "can't update config",
				Offered: supportedFormats,
			})
		}
	}
}
