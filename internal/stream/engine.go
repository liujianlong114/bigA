package stream

import (
	"context"
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/market"
)

// Engine 交易时段每秒拉全市场行情 → Redis → WS 广播 → 条件单检查
type Engine struct {
	universe    *market.Universe
	fetcher     *market.BatchFetcher
	redis       *cache.Redis
	hub         *Hub
	relax       bool
	interval    time.Duration
	seq         atomic.Int64
	busy        atomic.Bool
	conditional ConditionalChecker
}

// ConditionalChecker 条件单触发检查
type ConditionalChecker interface {
	CheckTriggers(ctx context.Context, quotes []market.LiveQuote)
}

func NewEngine(u *market.Universe, r *cache.Redis, h *Hub, relax bool) *Engine {
	return &Engine{
		universe: u,
		fetcher:  market.NewBatchFetcher(),
		redis:    r,
		hub:      h,
		relax:    relax,
		interval: time.Second,
	}
}

func (e *Engine) SetConditionalChecker(c ConditionalChecker) {
	e.conditional = c
}

func (e *Engine) Start(ctx context.Context) {
	go e.loop(ctx)
	log.Printf("stream: 实时行情引擎已启动 interval=%s relax=%v universe=%d", e.interval, e.relax, e.universe.Size())
}

func (e *Engine) Hub() *Hub { return e.hub }

func (e *Engine) loop(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.tick(ctx)
		}
	}
}

func (e *Engine) tick(ctx context.Context) {
	if e.universe.Size() == 0 {
		return
	}
	if !market.SessionOpen(time.Now(), e.relax) {
		return
	}
	if !e.busy.CompareAndSwap(false, true) {
		return
	}
	defer e.busy.Store(false)

	start := time.Now()
	quotes, err := e.fetcher.FetchAll(ctx, e.universe)
	if err != nil {
		log.Printf("stream: fetch error: %v", err)
		return
	}
	seq := e.seq.Add(1)
	updatedAt := time.Now().UnixMilli()

	redisMap := make(map[string][]byte, len(quotes))
	items := make([]any, 0, len(quotes))
	for _, q := range quotes {
		b, err := json.Marshal(q)
		if err != nil {
			continue
		}
		redisMap[q.Code] = b
		items = append(items, q)
	}

	meta := cache.LiveMeta{
		Seq: seq, UpdatedAt: updatedAt, Count: len(quotes),
		MarketOpen: true, Source: "sina_batch", DurationMs: time.Since(start).Milliseconds(),
	}
	if err := e.redis.StoreLiveQuotes(ctx, redisMap, meta); err != nil {
		log.Printf("stream: redis store: %v", err)
	} else if seq == 1 || seq%30 == 0 {
		log.Printf("stream: tick seq=%d count=%d duration=%dms", seq, len(quotes), meta.DurationMs)
	}

	for _, msg := range BuildQuoteChunks(seq, updatedAt, true, items) {
		e.hub.Broadcast(msg)
	}

	if e.conditional != nil {
		e.conditional.CheckTriggers(ctx, quotes)
	}
}

func (e *Engine) ReloadUniverse(ctx context.Context, rdb *cache.Redis, sources ...market.StockLister) error {
	return e.universe.ReloadUniverse(ctx, rdb, sources...)
}
