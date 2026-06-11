package market

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/store"
)

// StockRow 分页行情行
type StockRow struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Board     string  `json:"board"`
	BoardName string  `json:"board_name"`
	Price     float64 `json:"price"`
	ChangePct float64 `json:"change_pct"`
	PrevClose float64 `json:"prev_close"`
	Volume    int64   `json:"volume"`
	Amount    float64 `json:"amount"`
	LimitUp   float64 `json:"limit_up"`
	LimitDown float64 `json:"limit_down"`
}

type StockListResult struct {
	Total   int64      `json:"total"`
	Page    int        `json:"page"`
	Size    int        `json:"size"`
	Source  string     `json:"source"`
	Items   []StockRow `json:"items"`
	Updated int64      `json:"updated_at"`
}

var boardNames = map[string]string{
	"main": "主板", "chinext": "创业板", "star": "科创板", "st": "ST", "bj": "北交所",
}

// Catalog 全市场分页行情（Redis 实时 + MySQL 快照兜底）
type Catalog struct {
	redis    *cache.Redis
	repo     *store.Repo
	universe *Universe

	mu       sync.RWMutex
	cached   []StockRow
	cachedAt time.Time
}

func NewCatalog(r *cache.Redis, repo *store.Repo, u *Universe) *Catalog {
	return &Catalog{redis: r, repo: repo, universe: u}
}

func (c *Catalog) List(ctx context.Context, page, pageSize int, board, sortBy, keyword string) (*StockListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	all, source, updated := c.loadAll(ctx)
	if len(all) == 0 {
		return &StockListResult{Page: page, Size: pageSize, Source: source, Items: []StockRow{}}, nil
	}

	filtered := filterRows(all, board, keyword)
	sortRows(filtered, sortBy)

	total := int64(len(filtered))
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return &StockListResult{Total: total, Page: page, Size: pageSize, Source: source, Updated: updated, Items: []StockRow{}}, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	items := make([]StockRow, end-start)
	copy(items, filtered[start:end])
	return &StockListResult{
		Total: total, Page: page, Size: pageSize, Source: source, Updated: updated, Items: items,
	}, nil
}

func (c *Catalog) BoardStats(ctx context.Context) (map[string]any, error) {
	all, _, _ := c.loadAll(ctx)
	stats := map[string]int64{"all": int64(len(all))}
	for _, r := range all {
		stats[r.Board]++
	}
	names := map[string]string{"all": "全部"}
	for k, v := range boardNames {
		names[k] = v
	}
	return map[string]any{"counts": stats, "names": names}, nil
}

func (c *Catalog) Search(ctx context.Context, keyword string, limit int) ([]StockRow, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	all, _, _ := c.loadAll(ctx)
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw == "" {
		return nil, nil
	}
	var out []StockRow
	for _, r := range all {
		if strings.Contains(r.Code, keyword) || strings.Contains(strings.ToLower(r.Name), kw) {
			out = append(out, r)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (c *Catalog) loadAll(ctx context.Context) ([]StockRow, string, int64) {
	c.mu.RLock()
	if len(c.cached) > 0 && time.Since(c.cachedAt) < 3*time.Second {
		out := make([]StockRow, len(c.cached))
		copy(out, c.cached)
		c.mu.RUnlock()
		return out, "redis_cache", time.Now().UnixMilli()
	}
	c.mu.RUnlock()

	rows, source, updated := c.loadFromRedis(ctx)
	if len(rows) == 0 {
		rows, source = c.loadFromDB(ctx)
		updated = time.Now().UnixMilli()
	}
	if len(rows) == 0 && c.universe != nil {
		rows = c.loadFromUniverse()
		source = "universe"
		updated = time.Now().UnixMilli()
	}

	c.mu.Lock()
	c.cached = rows
	c.cachedAt = time.Now()
	c.mu.Unlock()
	return rows, source, updated
}

func (c *Catalog) loadFromRedis(ctx context.Context) ([]StockRow, string, int64) {
	if c.redis == nil {
		return nil, "", 0
	}
	meta, _ := c.redis.GetLiveMeta(ctx)
	var updated int64
	if meta != nil {
		updated = meta.UpdatedAt
	}
	raw, err := c.redis.Client().HGetAll(ctx, cache.LiveQuotesHash).Result()
	if err != nil || len(raw) == 0 {
		return nil, "", updated
	}
	rows := make([]StockRow, 0, len(raw))
	for code, payload := range raw {
		var q LiveQuote
		if err := json.Unmarshal([]byte(payload), &q); err != nil {
			continue
		}
		if q.Code == "" {
			q.Code = code
		}
		if q.Name == "" && c.universe != nil {
			q.Name = c.universe.Name(q.Code)
		}
		board := q.Board
		if board == "" {
			board = string(DetectBoard(q.Code, q.Name))
		}
		rows = append(rows, StockRow{
			Code: q.Code, Name: q.Name, Board: board, BoardName: boardNames[board],
			Price: q.Price, ChangePct: q.ChangePct, PrevClose: q.PrevClose,
			Volume: q.Volume, Amount: q.Amount, LimitUp: q.LimitUp, LimitDown: q.LimitDown,
		})
	}
	return rows, "redis_live", updated
}

func (c *Catalog) loadFromDB(ctx context.Context) ([]StockRow, string) {
	if c.repo == nil {
		return nil, ""
	}
	items, _, err := c.repo.ListStockSnapshots(ctx, "", "", "change_pct", 1, 10000)
	if err != nil || len(items) == 0 {
		return nil, ""
	}
	rows := make([]StockRow, len(items))
	for i, s := range items {
		rows[i] = snapshotToRow(s)
	}
	return rows, "mysql_snapshot"
}

func (c *Catalog) loadFromUniverse() []StockRow {
	if c.universe == nil {
		return nil
	}
	codes := c.universe.Codes()
	rows := make([]StockRow, 0, len(codes))
	for _, code := range codes {
		name := c.universe.Name(code)
		board := string(DetectBoard(code, name))
		rows = append(rows, StockRow{
			Code: code, Name: name, Board: board, BoardName: boardNames[board],
		})
	}
	return rows
}

func (c *Catalog) PersistSnapshots(ctx context.Context, quotes []LiveQuote) error {
	if c.repo == nil || len(quotes) == 0 {
		return nil
	}
	rows := make([]store.StockSnapshot, 0, len(quotes))
	now := time.Now()
	for _, q := range quotes {
		board := q.Board
		if board == "" {
			board = string(DetectBoard(q.Code, q.Name))
		}
		rows = append(rows, store.StockSnapshot{
			Code: q.Code, Name: q.Name, Board: board,
			Price: q.Price, ChangePct: q.ChangePct, PrevClose: q.PrevClose,
			Volume: q.Volume, Amount: q.Amount, LimitUp: q.LimitUp, LimitDown: q.LimitDown,
			UpdatedAt: now,
		})
	}
	return c.repo.BatchUpsertStockSnapshots(ctx, rows)
}

func snapshotToRow(s store.StockSnapshot) StockRow {
	return StockRow{
		Code: s.Code, Name: s.Name, Board: s.Board, BoardName: boardNames[s.Board],
		Price: s.Price, ChangePct: s.ChangePct, PrevClose: s.PrevClose,
		Volume: s.Volume, Amount: s.Amount, LimitUp: s.LimitUp, LimitDown: s.LimitDown,
	}
}

func filterRows(all []StockRow, board, keyword string) []StockRow {
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if board == "" || board == "all" {
		board = ""
	}
	if board == "" && kw == "" {
		out := make([]StockRow, len(all))
		copy(out, all)
		return out
	}
	out := make([]StockRow, 0, len(all)/4+1)
	for _, r := range all {
		if board != "" && r.Board != board {
			continue
		}
		if kw != "" && !strings.Contains(r.Code, keyword) && !strings.Contains(strings.ToLower(r.Name), kw) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func sortRows(rows []StockRow, sortBy string) {
	switch sortBy {
	case "change_pct_asc":
		sort.Slice(rows, func(i, j int) bool { return rows[i].ChangePct < rows[j].ChangePct })
	case "volume":
		sort.Slice(rows, func(i, j int) bool { return rows[i].Volume > rows[j].Volume })
	case "code":
		sort.Slice(rows, func(i, j int) bool { return rows[i].Code < rows[j].Code })
	case "name":
		sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	default:
		sort.Slice(rows, func(i, j int) bool { return rows[i].ChangePct > rows[j].ChangePct })
	}
}

func BoardLabel(board string) string {
	if v, ok := boardNames[board]; ok {
		return v
	}
	return board
}
