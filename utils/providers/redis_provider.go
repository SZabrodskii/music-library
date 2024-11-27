package providers

import (
	"context"
	"fmt"
	"github.com/SZabrodskii/music-library/utils"
	"github.com/go-redis/redis/v8"
)

type RedisProviderConfig struct {
	Addr string
}

func NewRedisProviderConfig() *RedisProviderConfig {
	return &RedisProviderConfig{
		Addr: utils.GetEnv("REDIS_URL", "redis:6379"),
	}
}

type RedisProvider struct {
	*redis.Client
}

func NewRedisProvider(config *RedisProviderConfig) (*RedisProvider, error) {
	client := redis.NewClient(&redis.Options{
		Addr: config.Addr,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisProvider{
		Client: client,
	}, nil
}

func (r *RedisProvider) GetClient() *redis.Client {
	return r.Client
}
