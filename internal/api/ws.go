package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/stream"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 64 * 1024,
}

type WSHandler struct {
	Hub   *stream.Hub
	Redis *cache.Redis
}

func (w *WSHandler) MarketQuotes(rw http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(rw, r, nil)
	if err != nil {
		return
	}
	client := stream.NewClient()
	w.Hub.Register(client)
	defer func() {
		w.Hub.Unregister(client)
		conn.Close()
	}()

	go w.pushSnapshot(r.Context(), client)

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for msg := range client.SendChan() {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

func (w *WSHandler) pushSnapshot(ctx context.Context, client *stream.Client) {
	meta, err := w.Redis.GetLiveMeta(ctx)
	if err != nil {
		return
	}
	var items []any
	cursor := uint64(0)
	for {
		res, next, err := w.Redis.Client().HScan(ctx, cache.LiveQuotesHash, cursor, "*", 500).Result()
		if err != nil {
			return
		}
		for i := 1; i < len(res); i += 2 {
			var q market.LiveQuote
			if json.Unmarshal([]byte(res[i]), &q) == nil {
				items = append(items, q)
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	if len(items) == 0 {
		metaMsg, _ := json.Marshal(stream.TickMessage{
			Type: "snapshot_meta", Seq: meta.Seq, UpdatedAt: meta.UpdatedAt,
			MarketOpen: meta.MarketOpen, Count: meta.Count,
		})
		client.TrySend(metaMsg)
		return
	}
	for _, msg := range stream.BuildQuoteChunks(meta.Seq, meta.UpdatedAt, meta.MarketOpen, items) {
		client.TrySend(msg)
	}
}

func (h *Handlers) LiveMeta(rw http.ResponseWriter, r *http.Request) {
	if h.LiveRedis == nil {
		writeErr(rw, http.StatusServiceUnavailable, "live stream not enabled")
		return
	}
	meta, err := h.LiveRedis.GetLiveMeta(r.Context())
	if err != nil {
		writeErr(rw, http.StatusNotFound, "no live data yet")
		return
	}
	writeJSON(rw, http.StatusOK, meta)
}

func (h *Handlers) LiveQuote(rw http.ResponseWriter, r *http.Request) {
	if h.LiveRedis == nil {
		writeErr(rw, http.StatusServiceUnavailable, "live stream not enabled")
		return
	}
	code := r.URL.Query().Get("code")
	if code != "" {
		b, err := h.LiveRedis.GetLiveQuote(r.Context(), code)
		if err != nil {
			writeErr(rw, http.StatusNotFound, "code not in live cache")
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		_, _ = rw.Write(b)
		return
	}
	cursor := uint64(0)
	if c := r.URL.Query().Get("cursor"); c != "" {
		n, err := strconv.ParseUint(c, 10, 64)
		if err == nil {
			cursor = n
		}
	}
	count := int64(200)
	res, next, err := h.LiveRedis.Client().HScan(r.Context(), cache.LiveQuotesHash, cursor, "*", count).Result()
	if err != nil {
		writeErr(rw, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]json.RawMessage, 0, len(res)/2)
	for i := 1; i < len(res); i += 2 {
		items = append(items, json.RawMessage(res[i]))
	}
	writeJSON(rw, http.StatusOK, map[string]any{
		"cursor": next, "items": items, "count": len(items),
	})
}
