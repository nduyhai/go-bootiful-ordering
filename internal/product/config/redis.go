package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

// RedisConfig holds the configuration for the Redis connection
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewDefaultRedisConfig creates a new RedisConfig with default values
func NewDefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	}
}

// Addr returns the address for the Redis connection
func (c *RedisConfig) Addr() string {
	return c.Host + ":" + c.Port
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr(),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}