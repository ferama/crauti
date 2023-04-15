package api

import (
	"net/http"
	"sync"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type configGroup struct {
	mu sync.Mutex
}

// Routes setup the root api routes
func configRoutes(router *gin.RouterGroup) {
	r := &configGroup{}

	router.GET("", r.config)
	router.GET("/writeable", r.writeable)
}

func (r *configGroup) config(c *gin.Context) {
	c.Negotiate(http.StatusOK, gin.Negotiate{
		Data:    conf.ConfInst,
		Offered: supportedFormats,
	})
}

func (r *configGroup) writeable(c *gin.Context) {
	r.mu.Lock()
	defer r.mu.Unlock()

	err := viper.WriteConfig()

	var res struct {
		Writeable bool `json:"writeable"`
	}
	res.Writeable = true
	if err != nil {
		res.Writeable = false
	}

	c.JSON(http.StatusOK, res)
}
