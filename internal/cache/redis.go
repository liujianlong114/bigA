package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const quoteKeyPrefix = "biga:quote:"
const quoteTTL = 5 * time.Second

type Redis struct {
	client *redis.Client
}

func New(addr, password string, db int) (*Redis, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Redis{client: c}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func quoteKey(code string) string {
	return quoteKeyPrefix + code
}

func (r *Redis) GetJSON(ctx context.Context, code string, dest any) bool {
	b, err := r.client.Get(ctx, quoteKey(code)).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(b, dest) == nil
}

func (r *Redis) SetJSON(ctx context.Context, code string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, quoteKey(code), b, quoteTTL).Err()
}

func (r *Redis) GetCacheJSON(ctx context.Context, key string, dest any) bool {
	b, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return false
	}
	return json.Unmarshal(b, dest) == nil
}

func (r *Redis) SetCacheJSON(ctx context.Context, key string, v any, ttl time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, b, ttl).Err()
}

func (r *Redis) LockTrade(ctx context.Context, accountID int64, ttl time.Duration) (func(), error) {
	key := fmt.Sprintf("biga:lock:account:%d", accountID)
	ok, err := r.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("account %d is busy", accountID)
	}
	return func() { _ = r.client.Del(context.Background(), key) }, nil
}
