package providers

import (
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisProvider struct {
	Logger *zap.Logger
	Client *redis.Client
}

func NewRedisProvider(logger *zap.Logger) *RedisProvider {
	client := redis.NewClient(&redis.Options{
		Addr: config.GetEnv("REDIS_URL"),
	})
	return &RedisProvider{
		Logger: logger,
		Client: client,
	}
}

func (r *RedisProvider) GetClient() *redis.Client {
	return r.Client
}
