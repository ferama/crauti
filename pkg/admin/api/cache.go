package api

import (
	"net/http"

	"github.com/ferama/crauti/pkg/redis"
	"github.com/gin-gonic/gin"
)

type cacheGroup struct{}

// cacheRoutes setup the root api routes
func cacheRoutes(router *gin.RouterGroup) {
	r := &cacheGroup{}

	router.POST("flush", r.flush)
	router.POST("flushall", r.flushAll)
}

func (r *cacheGroup) flushAll(c *gin.Context) {
	redis.CacheInstance().FlushallAsync()
	c.JSON(200, gin.H{
		"message": "full cache flush requested",
	})
}

// curl -X POST -d '{"match": "GET/api/config*"}' http://localhost:9000/cache/flush
func (r *cacheGroup) flush(c *gin.Context) {
	type mapping struct {
		Match string `json:"match"`
	}
	data := mapping{}
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	go redis.CacheInstance().Flush(data.Match)

	c.JSON(200, gin.H{
		"message": "cache flush requested",
	})
}
