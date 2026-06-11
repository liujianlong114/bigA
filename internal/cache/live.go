package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	LiveQuotesHash = "biga:live:quotes"
	LiveMetaKey    = "biga:live:meta"
	LivePubSub     = "biga:live:tick"
	LiveTTL        = 120 * time.Second
)

type LiveMeta struct {
	Seq        int64  `json:"seq"`
	UpdatedAt  int64  `json:"updated_at"`
	Count      int    `json:"count"`
	MarketOpen bool   `json:"market_open"`
	Source     string `json:"source"`
	DurationMs int64  `json:"duration_ms"`
}

// StoreLiveQuotes 批量写入 Redis Hash（全市场）
func (r *Redis) StoreLiveQuotes(ctx context.Context, quotes map[string][]byte, meta LiveMeta) error {
	pipe := r.client.Pipeline()
	for code, payload := range quotes {
		pipe.HSet(ctx, LiveQuotesHash, code, payload)
	}
	mb, _ := json.Marshal(meta)
	pipe.Set(ctx, LiveMetaKey, mb, LiveTTL)
	pipe.Expire(ctx, LiveQuotesHash, LiveTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	// pub/sub 广播 meta + 可选全量（客户端可据此拉 HGETALL 或等 WS 分片）
	tick, _ := json.Marshal(meta)
	return r.client.Publish(ctx, LivePubSub, tick).Err()
}

func (r *Redis) GetLiveMeta(ctx context.Context) (*LiveMeta, error) {
	b, err := r.client.Get(ctx, LiveMetaKey).Bytes()
	if err != nil {
		return nil, err
	}
	var m LiveMeta
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *Redis) GetLiveQuote(ctx context.Context, code string) ([]byte, error) {
	return r.client.HGet(ctx, LiveQuotesHash, code).Bytes()
}

func (r *Redis) GetLiveQuoteCount(ctx context.Context) (int64, error) {
	return r.client.HLen(ctx, LiveQuotesHash).Result()
}

func (r *Redis) SubscribeLive(ctx context.Context) *redis.PubSub {
	return r.client.Subscribe(ctx, LivePubSub)
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func LiveCodeKey(code string) string {
	return fmt.Sprintf("%s:%s", LiveQuotesHash, code)
}
