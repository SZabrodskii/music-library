package providers

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"time"
)

type CacheProvider struct {
	Logger     *zap.Logger
	Redis      *redis.Client
	localCache map[string]interface{}
}

func NewCacheProvider(logger *zap.Logger, redisProvider *RedisProvider) *CacheProvider {
	localCache := make(map[string]interface{})
	return &CacheProvider{
		Logger:     logger,
		Redis:      redisProvider.GetClient(),
		localCache: localCache,
	}
}

func (c *CacheProvider) GetFromCache(key string) (interface{}, bool) {
	if val, ok := c.localCache[key]; ok {
		return val, true
	}
	ctx := context.Background()
	val, err := c.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		c.Logger.Error("Failed to get from Redis", zap.Error(err))
		return nil, false
	}
	return val, true
}

func (c *CacheProvider) SetToCache(key string, value interface{}, ttl time.Duration) {
	c.localCache[key] = value
	ctx := context.Background()

	err := c.Redis.Set(ctx, key, value, ttl).Err()
	if err != nil {
		c.Logger.Error("Failed to set to Redis", zap.Error(err))
	}
}

func (c *CacheProvider) DeleteFromCache(key string) {
	delete(c.localCache, key)
	ctx := context.Background()

	err := c.Redis.Del(ctx, key).Err()
	if err != nil {
		c.Logger.Error("Failed to delete from Redis", zap.Error(err))
	}
}

func (c *CacheProvider) ClearCache() {
	c.localCache = make(map[string]interface{})
	ctx := context.Background()

	err := c.Redis.FlushDB(ctx).Err()
	if err != nil {
		c.Logger.Error("Failed to clear Redis", zap.Error(err))
	}
}
