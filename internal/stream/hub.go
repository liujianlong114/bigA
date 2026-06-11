package stream

import (
	"encoding/json"
	"sync"
)

// Hub WebSocket 广播中心
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*Client]struct{})}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	c.Close()
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		c.TrySend(msg)
	}
}

// Client WS 客户端
type Client struct {
	send chan []byte
	once sync.Once
}

func NewClient() *Client {
	return &Client{send: make(chan []byte, 64)}
}

func (c *Client) TrySend(msg []byte) {
	select {
	case c.send <- msg:
	default:
	}
}

func (c *Client) SendChan() <-chan []byte { return c.send }

func (c *Client) Close() {
	c.once.Do(func() { close(c.send) })
}

// TickMessage WS 推送消息
type TickMessage struct {
	Type       string          `json:"type"`
	Seq        int64           `json:"seq"`
	UpdatedAt  int64           `json:"updated_at"`
	MarketOpen bool            `json:"market_open"`
	Count      int             `json:"count"`
	Part       int             `json:"part,omitempty"`
	Total      int             `json:"total,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
}

const wsChunkSize = 500

func BuildQuoteChunks(seq, updatedAt int64, marketOpen bool, items []any) [][]byte {
	total := (len(items) + wsChunkSize - 1) / wsChunkSize
	meta, _ := json.Marshal(TickMessage{
		Type: "meta", Seq: seq, UpdatedAt: updatedAt, MarketOpen: marketOpen, Count: len(items),
	})
	out := [][]byte{meta}
	for i := 0; i < len(items); i += wsChunkSize {
		end := i + wsChunkSize
		if end > len(items) {
			end = len(items)
		}
		part := (i / wsChunkSize) + 1
		chunk, _ := json.Marshal(items[i:end])
		msg, _ := json.Marshal(TickMessage{
			Type: "quotes", Seq: seq, UpdatedAt: updatedAt, MarketOpen: marketOpen,
			Count: len(items), Part: part, Total: total, Data: chunk,
		})
		out = append(out, msg)
	}
	return out
}
