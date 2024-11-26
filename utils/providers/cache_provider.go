package providers

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"sync"
	"time"
)

type CacheProvider struct {
	logger     *zap.Logger
	redis      *redis.Client
	localCache map[string]interface{}
	mu         sync.RWMutex
}

func NewCacheProvider(logger *zap.Logger, redisProvider *RedisProvider) *CacheProvider {
	localCache := make(map[string]interface{})
	return &CacheProvider{
		logger:     logger,
		redis:      redisProvider.GetClient(),
		localCache: localCache,
	}
}

func (c *CacheProvider) GetFromCache(key string) (interface{}, bool) {
	c.mu.RLock()
	if val, ok := c.localCache[key]; ok {
		c.mu.RUnlock()
		return val, true
	}
	c.mu.RUnlock()

	ctx := context.Background()
	val, err := c.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	} else if err != nil {
		errors.Is(err, redis.Nil)
		return nil, false
	}
	c.mu.Lock()
	c.localCache[key] = val
	c.mu.Unlock()

	return val, true
}

func (c *CacheProvider) SetToCache(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	c.localCache[key] = value
	c.mu.Unlock()

	ctx := context.Background()

	if err := c.redis.Set(ctx, key, value, ttl).Err(); err != nil {
		c.logger.Error("Failed to set to Redis", zap.Error(err))
	} else {
		c.logger.Info("Key set successfully in Redis", zap.String("key", key))
	}
}

func (c *CacheProvider) DeleteFromCache(key string) {
	c.mu.Lock()
	if _, ok := c.localCache[key]; ok {
		delete(c.localCache, key)
	}
	c.mu.Unlock()

	ctx := context.Background()
	err := c.redis.Del(ctx, key).Err()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.logger.Warn("Key not found in Redis during deletion", zap.String("key", key))
			return
		}
		c.logger.Error("Failed to delete from Redis", zap.Error(err))
	} else {
		c.logger.Info("Key deleted successfully from Redis", zap.String("key", key))
	}
}

func (c *CacheProvider) ClearCache() {
	c.mu.Lock()
	c.localCache = make(map[string]interface{})
	c.mu.Unlock()

	ctx := context.Background()

	if err := c.redis.FlushDB(ctx).Err(); err != nil {
		c.logger.Error("Failed to clear Redis", zap.Error(err))
	} else {
		c.logger.Info("Redis cache cleared successfully")
	}
}
