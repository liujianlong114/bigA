package market

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// LiveQuote 实时行情（WS / Redis 用，字段精简）
type LiveQuote struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	PrevClose float64 `json:"prev_close"`
	ChangePct float64 `json:"change_pct"`
	Volume    int64   `json:"volume"`
	Amount    float64 `json:"amount"`
	Bid1      float64 `json:"bid1"`
	Ask1      float64 `json:"ask1"`
	Board     string  `json:"board,omitempty"`
	LimitUp   float64 `json:"limit_up,omitempty"`
	LimitDown float64 `json:"limit_down,omitempty"`
	Turnover  float64 `json:"turnover,omitempty"`
	UpdatedAt int64   `json:"updated_at"` // unix ms
}

// BatchFetcher 批量拉取全市场行情
type BatchFetcher struct {
	client *http.Client
}

func NewBatchFetcher() *BatchFetcher {
	return &BatchFetcher{client: &http.Client{Timeout: 45 * time.Second}}
}

const batchSize = 200

// FetchAll 按批拉取新浪全量行情（并发 4 路）
func (b *BatchFetcher) FetchAll(ctx context.Context, universe *Universe) ([]LiveQuote, error) {
	symbols := universe.Symbols()
	codes := universe.Codes()
	if len(symbols) == 0 {
		return nil, fmt.Errorf("empty universe")
	}

	type batch struct {
		idx     int
		symbols []string
		codes   []string
	}
	var batches []batch
	for i := 0; i < len(symbols); i += batchSize {
		end := i + batchSize
		if end > len(symbols) {
			end = len(symbols)
		}
		batches = append(batches, batch{
			idx: len(batches), symbols: symbols[i:end], codes: codes[i:end],
		})
	}

	type result struct {
		idx int
		q   []LiveQuote
		err error
	}
	ch := make(chan result, len(batches))
	sem := make(chan struct{}, 4)
	for _, bt := range batches {
		bt := bt
		go func() {
			sem <- struct{}{}
			part, err := b.fetchBatch(ctx, bt.symbols, bt.codes, universe)
			<-sem
			ch <- result{idx: bt.idx, q: part, err: err}
		}()
	}

	parts := make([][]LiveQuote, len(batches))
	var firstErr error
	for range batches {
		r := <-ch
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		bi := r.idx
		if bi < len(parts) {
			parts[bi] = r.q
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}
	var out []LiveQuote
	for _, p := range parts {
		out = append(out, p...)
	}
	return out, nil
}

func (b *BatchFetcher) fetchBatch(ctx context.Context, symbols, codes []string, u *Universe) ([]LiveQuote, error) {
	url := "https://hq.sinajs.cn/list=" + strings.Join(symbols, ",")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	utf8, _, _ := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), body)
	lines := strings.Split(string(utf8), "\n")
	now := time.Now().UnixMilli()
	var out []LiveQuote
	for i, line := range lines {
		if i >= len(codes) {
			break
		}
		name := u.Name(codes[i])
		q, ok := parseSinaLine(codes[i], name, line)
		if !ok {
			continue
		}
		board := DetectBoard(codes[i], name)
		q.Board = string(board)
		q.LimitUp, q.LimitDown = CalcLimitPrices(q.PrevClose, board)
		if q.PrevClose > 0 && q.Volume > 0 {
			// 估算换手率：成交量(手)*100 / 流通股本未知，用成交额/市值近似省略，此处用振幅代理
			q.Turnover = float64(q.Volume) / 10000 // 占位，异动模块主要用 volRatio
		}
		q.UpdatedAt = now
		out = append(out, q)
	}
	return out, nil
}

func parseSinaLine(code, name, line string) (LiveQuote, bool) {
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")
	if start < 0 || end <= start {
		return LiveQuote{}, false
	}
	fields := strings.Split(line[start+1:end], ",")
	if len(fields) < 10 {
		return LiveQuote{}, false
	}
	if name == "" {
		name = fields[0]
	}
	parse := func(i int) float64 {
		v, _ := strconv.ParseFloat(fields[i], 64)
		return v
	}
	open := parse(1)
	prev := parse(2)
	price := parse(3)
	high := parse(4)
	low := parse(5)
	vol, _ := strconv.ParseInt(fields[8], 10, 64)
	amount, _ := strconv.ParseFloat(fields[9], 64)
	pct := 0.0
	if prev > 0 {
		pct = (price - prev) / prev * 100
	}
	bid1, ask1 := 0.0, 0.0
	if len(fields) > 21 {
		bid1 = parse(11)
		ask1 = parse(21)
	}
	return LiveQuote{
		Code: code, Name: name, Price: price, Open: open, High: high, Low: low,
		PrevClose: prev, ChangePct: pct, Volume: vol / 100, Amount: amount,
		Bid1: bid1, Ask1: ask1,
	}, true
}
