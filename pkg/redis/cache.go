package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/redis/go-redis/v9"
)

var (
	once     sync.Once
	instance *cache
)

func CacheInstance() *cache {
	once.Do(func() {
		var red = conf.ConfInst.Redis
		instance = newCache(red.Host, red.Port, red.Password)
	})

	return instance
}

type cache struct {
	rdb *redis.Client
}

func newCache(host string, port int, password string) *cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       0, // use default DB
	})
	c := &cache{
		rdb: rdb,
	}
	ctx := context.Background()
	// run an unintrusive command to start the client connection
	rdb.Time(ctx)

	return c
}

func (c *cache) GetInt(key string) (int, error) {
	ctx := context.Background()
	val, err := c.rdb.Get(ctx, key).Int()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *cache) Get(key string) ([]byte, error) {
	ctx := context.Background()
	val, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (c *cache) Set(key string, body interface{}, ttl time.Duration) error {
	ctx := context.Background()
	err := c.rdb.Set(ctx, key, body, ttl)
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
