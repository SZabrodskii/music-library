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
	localCache map[string]*CacheItem
	mu         sync.RWMutex
}

func NewCacheProvider(logger *zap.Logger, redisProvider *RedisProvider) *CacheProvider {
	localCache := make(map[string]*CacheItem)

	cp := &CacheProvider{
		logger:     logger,
		redis:      redisProvider.GetClient(),
		localCache: localCache,
		mu:         sync.RWMutex{},
	}
	go cp.revalidateCache()
	return cp
}

type CacheItem struct {
	Body []byte
	TTL  time.Time
}

func (c *CacheProvider) revalidateCache() {
	for {
		time.Sleep(time.Minute)
		for key, val := range c.localCache {
			if val.TTL.After(time.Now()) {
				delete(c.localCache, key)
			}

		}
	}
}

func (c *CacheProvider) GetFromCache(key string) ([]byte, bool) {
	c.mu.RLock()
	if val, ok := c.localCache[key]; ok {
		c.mu.RUnlock()
		return val.Body, true
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
	c.localCache[key] = &CacheItem{
		Body: []byte(val),
		TTL:  time.Now().Add(time.Minute * 10),
	}
	c.mu.Unlock()

	return []byte(val), true
}

func (c *CacheProvider) SetToCache(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	c.localCache[key] = &CacheItem{
		Body: value,
		TTL:  time.Now().Add(ttl),
	}
	c.mu.Unlock()

	ctx := context.Background()

	if err := c.redis.Set(ctx, key, value, ttl).Err(); err != nil {
		c.logger.Error("Failed to set to Redis", zap.Error(err))
	} else {
		c.logger.Debug("Key set successfully in Redis", zap.String("key", key))
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
		c.logger.Debug("Key deleted successfully from Redis", zap.String("key", key))
	}
}

func (c *CacheProvider) ClearCache() {
	c.mu.Lock()
	c.localCache = make(map[string]*CacheItem)
	c.mu.Unlock()

	ctx := context.Background()

	if err := c.redis.FlushDB(ctx).Err(); err != nil {
		c.logger.Error("Failed to clear Redis", zap.Error(err))
	} else {
		c.logger.Debug("Redis cache cleared successfully")
	}
}
