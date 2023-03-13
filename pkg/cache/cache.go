package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/go-redis/redis/v8"
)

var red = conf.ConfInst.Middlewares.Cache.Redis
var CacheInst = newCache(red.Host, red.Port, red.Password)

type cache struct {
	rdb *redis.Client
}

func newCache(host string, port int, password string) *cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password, // no password set
		DB:       0,        // use default DB
	})
	c := &cache{
		rdb: rdb,
	}
	ctx := context.Background()
	// run an unintrusive command to start the client connection
	rdb.Time(ctx)

	return c
}

func (c *cache) Get(key string) ([]byte, error) {
	ctx := context.Background()
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c *cache) Set(key string, body []byte, ttl *time.Duration) error {
	ctx := context.Background()
	err := c.rdb.Set(ctx, key, body, *ttl)
	if err != nil {
		return err.Err()
	}
	return nil
}

// FlushallAsync (useful for tests)
func (c *cache) Flushall() error {
	ctx := context.Background()
	err := c.rdb.FlushAll(ctx)
	if err != nil {
		return err.Err()
	}
	return nil
}

// FlushallAsync deletes all cache keys
func (c *cache) FlushallAsync() error {
	ctx := context.Background()
	err := c.rdb.FlushAllAsync(ctx)
	if err != nil {
		return err.Err()
	}
	return nil
}

// Flush all keys matching pattern
func (c *cache) Flush(match string) (int, error) {
	ctx := context.Background()
	var cursor uint64 = 0
	flushedKeys := 0
	for {
		var keys []string
		var err error
		keys, cursor, err = c.rdb.Scan(ctx, cursor, match, 0).Result()
		if err != nil {
			return flushedKeys, err
		}

		for _, key := range keys {
			// I'm using expire with 0 here because it's complexity as per docs
			// is O(1) while del is O(n)
			c.rdb.Expire(ctx, key, 0)
			flushedKeys++
		}

		// no more keys
		if cursor == 0 {
			break
		}
	}
	return flushedKeys, nil
}
