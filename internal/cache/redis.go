package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient обертка для Redis клиента
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient создает новый Redis клиент
func NewRedisClient() *RedisClient {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	db := 0

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
		PoolSize: 20,
	})

	return &RedisClient{client: client}
}

// Ping проверяет подключение к Redis
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close закрывает подключение к Redis
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Set сохраняет значение в Redis
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Get получает значение из Redis
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return json.Unmarshal(data, dest)
}

// Increment увеличивает счетчик
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

// GetMetrics возвращает метрики Redis
func (r *RedisClient) GetMetrics(ctx context.Context) (map[string]interface{}, error) {
	info, err := r.client.Info(ctx).Result()
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]interface{})
	metrics["info"] = info

	// Получаем статистику ключей
	keys, _ := r.client.Keys(ctx, "*").Result()
	metrics["total_keys"] = len(keys)

	// Получаем использование памяти
	memory, _ := r.client.Info(ctx, "memory").Result()
	metrics["memory_info"] = memory

	return metrics, nil
}

// GetClient возвращает оригинальный Redis клиент
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}
