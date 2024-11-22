package cache

import (
	"context"
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

type Cache struct {
	Logger     *zap.Logger
	rdb        *redis.Client
	localCache map[string]interface{}
}

func NewCache(logger *zap.Logger) *Cache {
	localCache := make(map[string]interface{})
	rdb := redis.NewClient(&redis.Options{
		Addr: config.GetEnv("REDIS_URL"),
	})
	return &Cache{
		Logger:     logger,
		rdb:        rdb,
		localCache: localCache,
	}
}

func (c *Cache) GetFromCache(key string) (interface{}, bool) {
	if val, ok := c.localCache[key]; ok {
		return val, true
	}
	ctx := context.Background()
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		c.Logger.Error("Failed to get from Redis", zap.Error(err))
		return nil, false
	}
	return val, true
}

func (c *Cache) SetToCache(key string, value interface{}, ttl time.Duration) {
	c.localCache[key] = value
	ctx := context.Background()

	err := c.rdb.Set(ctx, key, value, ttl).Err()
	if err != nil {
		c.Logger.Error("Failed to set to Redis", zap.Error(err))
	}
}

func (c *Cache) DeleteFromCache(key string) {
	delete(c.localCache, key)
	ctx := context.Background()

	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		c.Logger.Error("Failed to delete from Redis", zap.Error(err))
	}
}

func (c *Cache) ClearCache() {
	c.localCache = make(map[string]interface{})
	ctx := context.Background()

	err := c.rdb.FlushDB(ctx).Err()
	if err != nil {
		c.Logger.Error("Failed to clear Redis", zap.Error(err))
	}
}
