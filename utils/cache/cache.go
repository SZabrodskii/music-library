package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"music-library/gateway/config"
	"time"
)

var logger *zap.Logger
var rdb *redis.Client
var localCache map[string]interface{}

func NewCache() *Cache {
	logger, _ = zap.NewProduction()
	localCache = make(map[string]interface{})
	rdb = redis.NewClient(&redis.Options{
		Addr: config.GetEnv("REDIS_URL"),
	})
	return &Cache{
		Logger:     logger,
		rdb:        rdb,
		localCache: localCache,
	}
}

type Cache struct {
	Logger     *zap.Logger
	rdb        *redis.Client
	localCache map[string]interface{}
}

func (c *Cache) GetFromCache(key string) (interface{}, bool) {
	if val, ok := localCache[key]; ok {
		return val, true
	}
	ctx := context.Background()
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		logger.Error("Failed to get from Redis", zap.Error(err))
		return nil, false
	}
	return val, true
}

func (c *Cache) SetToCache(key string, value interface{}, ttl time.Duration) {
	localCache[key] = value
	ctx := context.Background()

	err := c.rdb.Set(ctx, key, value, ttl).Err()
	if err != nil {
		logger.Error("Failed to set to Redis", zap.Error(err))
	}

}

func (c *Cache) DeleteFromCache(key string) {
	delete(localCache, key)
	ctx := context.Background()

	err := c.rdb.Del(ctx, key).Err()
	if err != nil {
		logger.Error("Failed to delete from Redis", zap.Error(err))
	}
}

func (c *Cache) ClearCache() {
	localCache = make(map[string]interface{})
	ctx := context.Background()

	err := rdb.FlushDB(ctx).Err()
	if err != nil {
		logger.Error("Failed to clear Redis", zap.Error(err))
	}
}