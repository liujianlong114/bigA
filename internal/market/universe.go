package market

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lijianjun/bigA/internal/cache"
)

// Universe 全 A 股代码表
type Universe struct {
	mu      sync.RWMutex
	codes   []string
	symbols []string
	names   map[string]string
}

func NewUniverse() *Universe {
	return &Universe{names: make(map[string]string)}
}

func (u *Universe) Codes() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	out := make([]string, len(u.codes))
	copy(out, u.codes)
	return out
}

func (u *Universe) Symbols() []string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	out := make([]string, len(u.symbols))
	copy(out, u.symbols)
	return out
}

func (u *Universe) Name(code string) string {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return u.names[code]
}

func (u *Universe) Size() int {
	u.mu.RLock()
	defer u.mu.RUnlock()
	return len(u.codes)
}

// StockLister 分页股票列表
type StockLister interface {
	List(ctx context.Context, page, pageSize int) ([]StockBrief, int, error)
}

// LoadAll 加载全 A 股：Redis 缓存 → 东方财富 → 新浪（带间隔）
func (u *Universe) LoadAll(ctx context.Context, rdb *cache.Redis, sources ...StockLister) error {
	if rdb != nil {
		if rows, err := rdb.LoadUniverse(ctx); err == nil && len(rows) > 1000 {
			u.applyRecords(rows)
			log.Printf("universe: 从 Redis 加载 %d 只", u.Size())
			return nil
		}
	}

	var lastErr error
	for i, src := range sources {
		if i > 0 {
			time.Sleep(2 * time.Second)
		}
		if err := u.loadPaginated(ctx, src, true); err != nil {
			lastErr = err
			log.Printf("universe: source %d failed: %v", i, err)
			continue
		}
		if u.Size() < 1000 {
			lastErr = fmt.Errorf("too few symbols: %d", u.Size())
			continue
		}
		if rdb != nil {
			_ = u.SaveToRedis(ctx, rdb)
		}
		log.Printf("universe: 在线加载 %d 只", u.Size())
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("universe: no sources")
}

func (u *Universe) loadPaginated(ctx context.Context, list StockLister, slow bool) error {
	const pageSize = 100
	var all []StockBrief
	page := 1
	total := 0
	for {
		if slow {
			time.Sleep(350 * time.Millisecond)
		}
		items, t, err := list.List(ctx, page, pageSize)
		if err != nil {
			return err
		}
		if total == 0 && t > 0 {
			total = t
		}
		if len(items) == 0 {
			break
		}
		all = append(all, items...)
		if total > 0 && len(all) >= total {
			break
		}
		if len(items) < pageSize {
			break
		}
		page++
		if page > 80 {
			break
		}
	}
	if len(all) == 0 {
		return fmt.Errorf("empty list")
	}
	u.applyBriefs(all)
	return nil
}

func (u *Universe) applyBriefs(all []StockBrief) {
	codes := make([]string, 0, len(all))
	symbols := make([]string, 0, len(all))
	names := make(map[string]string, len(all))
	for _, s := range all {
		if len(s.Code) != 6 {
			continue
		}
		sym := SinaSymbol(s.Code)
		if sym == "" {
			continue
		}
		codes = append(codes, s.Code)
		symbols = append(symbols, sym)
		names[s.Code] = s.Name
	}
	u.mu.Lock()
	u.codes = codes
	u.symbols = symbols
	u.names = names
	u.mu.Unlock()
}

func (u *Universe) applyRecords(rows []cache.UniverseRecord) {
	codes := make([]string, 0, len(rows))
	symbols := make([]string, 0, len(rows))
	names := make(map[string]string, len(rows))
	for _, row := range rows {
		if len(row.Code) != 6 {
			continue
		}
		sym := row.Symbol
		if sym == "" {
			sym = SinaSymbol(row.Code)
		}
		if sym == "" {
			continue
		}
		codes = append(codes, row.Code)
		symbols = append(symbols, sym)
		names[row.Code] = row.Name
	}
	u.mu.Lock()
	u.codes = codes
	u.symbols = symbols
	u.names = names
	u.mu.Unlock()
}

func (u *Universe) SaveToRedis(ctx context.Context, rdb *cache.Redis) error {
	u.mu.RLock()
	rows := make([]cache.UniverseRecord, len(u.codes))
	for i, code := range u.codes {
		rows[i] = cache.UniverseRecord{
			Code: code, Name: u.names[code], Symbol: u.symbols[i],
		}
	}
	u.mu.RUnlock()
	return rdb.SaveUniverse(ctx, rows)
}

// Load 兼容旧调用
func (u *Universe) Load(ctx context.Context, list *SinaList) error {
	return u.loadPaginated(ctx, list, true)
}

func (u *Universe) ReloadUniverse(ctx context.Context, rdb *cache.Redis, sources ...StockLister) error {
	u.mu.Lock()
	u.codes = nil
	u.symbols = nil
	u.names = make(map[string]string)
	u.mu.Unlock()
	return u.LoadAll(ctx, rdb, sources...)
}
