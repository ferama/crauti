package cache

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type cacheRoutes struct {
	cache *cache
}

// Routes setup the root api routes
func Routes(router *gin.RouterGroup) {
	r := &cacheRoutes{
		cache: CacheInst,
	}

	router.POST("flush", r.Flush)
	router.POST("flushall", r.FlushAll)
}

func (r *cacheRoutes) FlushAll(c *gin.Context) {
	r.cache.FlushallAsync()
	c.JSON(200, gin.H{
		"message": "full cache flush requested",
	})
}

// curl -X POST -d '{"match": "GET/api/config*"}' http://localhost:9000/cache/flush
func (r *cacheRoutes) Flush(c *gin.Context) {
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
	go r.cache.Flush(data.Match)

	c.JSON(200, gin.H{
		"message": "cache flush requested",
	})
}
