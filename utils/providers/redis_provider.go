package providers

import (
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"os"
)

type RedisProviderConfig struct {
	Addr string
}

func NewRedisProviderConfig() *RedisProviderConfig {
	return &RedisProviderConfig{
		Addr: os.Getenv("REDIS_URL"),
	}
}

type RedisProvider struct {
	logger *zap.Logger
	client *redis.Client
}

func NewRedisProvider(logger *zap.Logger, config *RedisProviderConfig) *RedisProvider {
	client := redis.NewClient(&redis.Options{
		Addr: config.Addr,
	})
	return &RedisProvider{
		logger: logger,
		client: client,
	}
}

func (r *RedisProvider) GetClient() *redis.Client {
	return r.client
}
