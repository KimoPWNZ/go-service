package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient обертка для клиента Redis
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient создает новый клиент Redis
func NewRedisClient() *RedisClient {
	host := getEnv("REDIS_HOST", "localhost")
	port := getEnv("REDIS_PORT", "6379")
	password := getEnv("REDIS_PASSWORD", "")

	addr := fmt.Sprintf("%s:%s", host, port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &RedisClient{client: rdb}
}

// Ping проверяет подключение к Redis
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Set сохраняет значение в Redis
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get получает значение из Redis
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Close закрывает соединение с Redis
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// getEnv получает переменную окружения с дефолтным значением
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Publish публикует сообщение в канал
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, data).Err()
}

// Subscribe подписывается на канал
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}
