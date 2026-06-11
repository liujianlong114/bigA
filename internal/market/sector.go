package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lijianjun/bigA/internal/cache"
)

// SectorKind 板块类型
type SectorKind string

const (
	SectorIndustry SectorKind = "industry" // 申万/行业
	SectorConcept  SectorKind = "concept"  // 概念
)

// SectorBrief 板块摘要
type SectorBrief struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	ChangePct  float64 `json:"change_pct"`
	MainInflow float64 `json:"main_inflow"`
	Amount     float64 `json:"amount"`
	Turnover   float64 `json:"turnover"`
}

// SectorListResult 板块列表
type SectorListResult struct {
	Type      string        `json:"type"`
	Total     int           `json:"total"`
	Items     []SectorBrief `json:"items"`
	Source    string        `json:"source"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// SectorStocksResult 板块成分股
type SectorStocksResult struct {
	SectorCode string       `json:"sector_code"`
	SectorName string       `json:"sector_name,omitempty"`
	Total      int          `json:"total"`
	Items      []StockBrief `json:"items"`
	Source     string       `json:"source"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

func (k SectorKind) fsFilter() string {
	switch k {
	case SectorConcept:
		return "m:90+t:3+f:!50"
	default:
		return "m:90+t:2+f:!50"
	}
}

func normalizeSectorKind(raw string) SectorKind {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "concept", "gn":
		return SectorConcept
	default:
		return SectorIndustry
	}
}

func normalizeSectorCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return ""
	}
	upper := strings.ToUpper(code)
	if strings.HasPrefix(upper, "BK") && len(upper) > 2 {
		return upper
	}
	// 新浪节点 ID 保持原样
	if strings.HasPrefix(code, "new_") || strings.HasPrefix(code, "sw") || strings.HasPrefix(code, "chgn_") {
		return code
	}
	return "BK" + upper
}

// ListSectors 申万行业 / 概念板块列表
func (e *EastMoney) ListSectors(ctx context.Context, kind SectorKind, page, pageSize int) (*SectorListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	params := url.Values{}
	params.Set("pn", strconv.Itoa(page))
	params.Set("pz", strconv.Itoa(pageSize))
	params.Set("po", "1")
	params.Set("np", "1")
	params.Set("ut", eastMoneyUT)
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("fid", "f3")
	params.Set("fs", kind.fsFilter())
	params.Set("fields", "f12,f14,f2,f3,f62,f184,f66,f204,f205,f206")

	body, err := e.doGet(ctx, eastMoneyListPath, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			Total int `json:"total"`
			Diff  []struct {
				Code       string  `json:"f12"`
				Name       string  `json:"f14"`
				Price      float64 `json:"f2"`
				ChangePct  float64 `json:"f3"`
				MainInflow float64 `json:"f62"`
				Turnover   float64 `json:"f184"`
				Amount     float64 `json:"f66"`
			} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("eastmoney sectors: %w", errJSONBody(body, err))
	}

	items := make([]SectorBrief, 0, len(resp.Data.Diff))
	for _, row := range resp.Data.Diff {
		if row.Code == "" || row.Name == "" {
			continue
		}
		items = append(items, SectorBrief{
			Code:       row.Code,
			Name:       row.Name,
			Price:      row.Price,
			ChangePct:  row.ChangePct,
			MainInflow: row.MainInflow,
			Amount:     row.Amount,
			Turnover:   row.Turnover,
		})
	}

	return &SectorListResult{
		Type:      string(kind),
		Total:     resp.Data.Total,
		Items:     items,
		Source:    "eastmoney",
		UpdatedAt: time.Now(),
	}, nil
}

// ListSectorStocks 板块成分股
func (e *EastMoney) ListSectorStocks(ctx context.Context, sectorCode string, page, pageSize int) (*SectorStocksResult, error) {
	sectorCode = normalizeSectorCode(sectorCode)
	if sectorCode == "" {
		return nil, fmt.Errorf("invalid sector code")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	params := url.Values{}
	params.Set("pn", strconv.Itoa(page))
	params.Set("pz", strconv.Itoa(pageSize))
	params.Set("po", "1")
	params.Set("np", "1")
	params.Set("ut", eastMoneyUT)
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("fid", "f3")
	params.Set("fs", "b:"+sectorCode)
	params.Set("fields", "f12,f14,f2,f3,f62,f184,f66")

	body, err := e.doGet(ctx, eastMoneyListPath, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			Total int `json:"total"`
			Diff  []struct {
				Code       string  `json:"f12"`
				Name       string  `json:"f14"`
				Price      float64 `json:"f2"`
				ChangePct  float64 `json:"f3"`
				MainInflow float64 `json:"f62"`
				Turnover   float64 `json:"f184"`
				Amount     float64 `json:"f66"`
			} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("eastmoney sectors: %w", errJSONBody(body, err))
	}

	items := make([]StockBrief, 0, len(resp.Data.Diff))
	for _, row := range resp.Data.Diff {
		if row.Code == "" {
			continue
		}
		items = append(items, StockBrief{
			Code:      row.Code,
			Name:      row.Name,
			Price:     row.Price,
			ChangePct: row.ChangePct,
		})
	}

	return &SectorStocksResult{
		SectorCode: sectorCode,
		Total:      resp.Data.Total,
		Items:      items,
		Source:     "eastmoney",
		UpdatedAt:  time.Now(),
	}, nil
}

// SectorService 板块行情（东财优先，新浪兜底，Redis 缓存）
type SectorService struct {
	em      *EastMoney
	sina    *SinaSector
	cache   *cache.Redis
	catalog *Catalog
}

func NewSectorService(em *EastMoney, c *cache.Redis) *SectorService {
	return &SectorService{em: em, sina: NewSinaSector(), cache: c}
}

func (s *SectorService) SetCatalog(c *Catalog) {
	s.catalog = c
}

func sectorCacheKey(kind SectorKind, page, pageSize int) string {
	return fmt.Sprintf("biga:sector:list:%s:%d:%d", kind, page, pageSize)
}

func sectorStaleKey(kind SectorKind, page, pageSize int) string {
	return fmt.Sprintf("biga:sector:list:lastgood:%s:%d:%d", kind, page, pageSize)
}

func emptySectorList(kind SectorKind) *SectorListResult {
	return &SectorListResult{
		Type:      string(kind),
		Total:     0,
		Items:     []SectorBrief{},
		Source:    "offline",
		UpdatedAt: time.Now(),
	}
}

func (s *SectorService) listSectors(ctx context.Context, kind SectorKind, page, pageSize int) (*SectorListResult, error) {
	res, err := s.em.ListSectors(ctx, kind, page, pageSize)
	if err == nil && len(res.Items) > 0 {
		return res, nil
	}
	res, sinaErr := s.sina.ListSectors(ctx, kind, page, pageSize)
	if sinaErr == nil && len(res.Items) > 0 {
		return res, nil
	}
	if err != nil {
		return nil, err
	}
	if sinaErr != nil {
		return nil, sinaErr
	}
	return res, nil
}

func (s *SectorService) loadSectorStale(ctx context.Context, kind SectorKind, page, pageSize int) *SectorListResult {
	if s.cache == nil {
		return nil
	}
	var cached SectorListResult
	if s.cache.GetCacheJSON(ctx, sectorStaleKey(kind, page, pageSize), &cached) && len(cached.Items) > 0 {
		return &cached
	}
	return nil
}

func (s *SectorService) leaderboardFallback(ctx context.Context, page, pageSize int) *SectorListResult {
	if s.catalog == nil {
		return emptySectorList(SectorIndustry)
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 10
	}
	res, err := s.catalog.List(ctx, page, pageSize, "all", "change_pct", "")
	if err != nil || len(res.Items) == 0 {
		return emptySectorList(SectorIndustry)
	}
	items := make([]SectorBrief, 0, len(res.Items))
	for _, row := range res.Items {
		items = append(items, SectorBrief{
			Code:      row.Code,
			Name:      row.Name,
			Price:     row.Price,
			ChangePct: row.ChangePct,
		})
	}
	return &SectorListResult{
		Type:      string(SectorIndustry),
		Total:     len(items),
		Items:     items,
		Source:    "leaderboard_fallback",
		UpdatedAt: time.Now(),
	}
}

func (s *SectorService) listSectorStocks(ctx context.Context, sectorCode string, page, pageSize int) (*SectorStocksResult, error) {
	sectorCode = normalizeSectorCode(sectorCode)
	// 新浪节点形如 new_hghy / chgn_xxx，BK 前缀走东财
	if strings.HasPrefix(sectorCode, "BK") {
		res, err := s.em.ListSectorStocks(ctx, sectorCode, page, pageSize)
		if err == nil && len(res.Items) > 0 {
			return res, nil
		}
	}
	code := strings.TrimPrefix(sectorCode, "BK")
	return s.sina.ListSectorStocks(ctx, code, page, pageSize)
}

func (s *SectorService) ListSectors(ctx context.Context, kind SectorKind, page, pageSize int) (*SectorListResult, error) {
	key := sectorCacheKey(kind, page, pageSize)
	var cached SectorListResult
	if s.cache != nil && s.cache.GetCacheJSON(ctx, key, &cached) && len(cached.Items) > 0 {
		return &cached, nil
	}
	res, err := s.listSectors(ctx, kind, page, pageSize)
	if err != nil || len(res.Items) == 0 {
		if stale := s.loadSectorStale(ctx, kind, page, pageSize); stale != nil {
			return stale, nil
		}
		if fb := s.leaderboardFallback(ctx, page, pageSize); len(fb.Items) > 0 {
			return fb, nil
		}
		return emptySectorList(kind), nil
	}
	if s.cache != nil {
		_ = s.cache.SetCacheJSON(ctx, key, res, 60*time.Second)
		_ = s.cache.SetCacheJSON(ctx, sectorStaleKey(kind, page, pageSize), res, 7*24*time.Hour)
	}
	return res, nil
}

func (s *SectorService) ListSectorStocks(ctx context.Context, sectorCode string, page, pageSize int) (*SectorStocksResult, error) {
	sectorCode = normalizeSectorCode(sectorCode)
	key := fmt.Sprintf("biga:sector:stocks:%s:%d:%d", sectorCode, page, pageSize)
	var cached SectorStocksResult
	if s.cache != nil && s.cache.GetCacheJSON(ctx, key, &cached) && len(cached.Items) > 0 {
		return &cached, nil
	}
	res, err := s.listSectorStocks(ctx, sectorCode, page, pageSize)
	if err != nil || len(res.Items) == 0 {
		// 领涨榜 fallback 的 code 是股票代码
		if s.catalog != nil && len(sectorCode) == 6 && !strings.HasPrefix(sectorCode, "BK") {
			if pageRes, listErr := s.catalog.List(ctx, 1, 1, "all", "code", sectorCode); listErr == nil && len(pageRes.Items) > 0 {
				row := pageRes.Items[0]
				return &SectorStocksResult{
					SectorCode: sectorCode,
					SectorName: row.Name,
					Total:      1,
					Items: []StockBrief{{
						Code: row.Code, Name: row.Name, Price: row.Price, ChangePct: row.ChangePct,
					}},
					Source:    "catalog",
					UpdatedAt: time.Now(),
				}, nil
			}
		}
		return &SectorStocksResult{
			SectorCode: sectorCode,
			Total:      0,
			Items:      []StockBrief{},
			Source:     "offline",
			UpdatedAt:  time.Now(),
		}, nil
	}
	if s.cache != nil {
		_ = s.cache.SetCacheJSON(ctx, key, res, 60*time.Second)
	}
	return res, nil
}
