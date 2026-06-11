package cache

import (
	"context"
	"encoding/json"
	"time"
)

const universeKey = "biga:universe:v1"

type UniverseRecord struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// SaveUniverse 缓存全市场代码表（避免频繁拉新浪列表）
func (r *Redis) SaveUniverse(ctx context.Context, rows []UniverseRecord) error {
	b, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, universeKey, b, 7*24*time.Hour).Err()
}

func (r *Redis) LoadUniverse(ctx context.Context) ([]UniverseRecord, error) {
	b, err := r.client.Get(ctx, universeKey).Bytes()
	if err != nil {
		return nil, err
	}
	var rows []UniverseRecord
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}
