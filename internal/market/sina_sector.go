package market

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const sinaSectorBase = "http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php"

// SinaSector 新浪行业/概念板块
type SinaSector struct {
	client *http.Client
}

func NewSinaSector() *SinaSector {
	return &SinaSector{client: &http.Client{Timeout: 15 * time.Second}}
}

type sinaSectorCatalog struct {
	Industry []sectorNode `json:"industry"`
	Concept  []sectorNode `json:"concept"`
	Updated  time.Time    `json:"updated"`
}

type sectorNode struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type sinaHQRow struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	Trade         string  `json:"trade"`
	PriceChange   float64 `json:"pricechange"`
	ChangePercent float64 `json:"changepercent"`
	Amount        float64 `json:"amount"`
	TurnoverRatio float64 `json:"turnoverratio"`
}

func (s *SinaSector) loadCatalog(ctx context.Context) (*sinaSectorCatalog, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sinaSectorBase+"/Market_Center.getHQNodes", nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}

	var root any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, errJSONBody(body, err)
	}

	cat := &sinaSectorCatalog{Updated: time.Now()}
	walkHQNodes(root, func(name, node string) {
		if strings.HasPrefix(node, "chgn_") {
			cat.Concept = append(cat.Concept, sectorNode{Code: node, Name: name})
			return
		}
		if strings.HasPrefix(node, "new_") || strings.HasPrefix(node, "sw") {
			cat.Industry = append(cat.Industry, sectorNode{Code: node, Name: name})
		}
	})
	if len(cat.Industry) == 0 && len(cat.Concept) == 0 {
		return nil, fmt.Errorf("sina: empty sector catalog")
	}
	return cat, nil
}

func walkHQNodes(v any, emit func(name, node string)) {
	switch t := v.(type) {
	case []any:
		if len(t) >= 3 {
			name, _ := t[0].(string)
			node, _ := t[2].(string)
			if name != "" && node != "" {
				emit(name, node)
			}
		}
		for _, child := range t {
			walkHQNodes(child, emit)
		}
	}
}

func (s *SinaSector) ListSectors(ctx context.Context, kind SectorKind, page, pageSize int) (*SectorListResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	cat, err := s.loadCatalog(ctx)
	if err != nil {
		return nil, err
	}

	nodes := cat.Industry
	if kind == SectorConcept {
		nodes = cat.Concept
	}
	total := len(nodes)
	start := (page - 1) * pageSize
	if start >= total {
		return &SectorListResult{
			Type: string(kind), Total: total, Items: nil, Source: "sina", UpdatedAt: time.Now(),
		}, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	slice := nodes[start:end]

	items := make([]SectorBrief, len(slice))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)
	for i, n := range slice {
		wg.Add(1)
		go func(i int, n sectorNode) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			items[i] = s.summarizeNode(ctx, n)
		}(i, n)
	}
	wg.Wait()

	return &SectorListResult{
		Type:      string(kind),
		Total:     total,
		Items:     items,
		Source:    "sina",
		UpdatedAt: time.Now(),
	}, nil
}

func (s *SinaSector) summarizeNode(ctx context.Context, n sectorNode) SectorBrief {
	rows, err := s.fetchNodeData(ctx, n.Code, 1, 80)
	brief := SectorBrief{Code: n.Code, Name: n.Name}
	if err != nil || len(rows) == 0 {
		return brief
	}
	var sumPct, sumAmt, sumTurn, sumPrice float64
	var cnt float64
	for _, r := range rows {
		if r.ChangePercent == 0 && r.Trade == "" {
			continue
		}
		price, _ := strconv.ParseFloat(r.Trade, 64)
		sumPct += r.ChangePercent
		sumAmt += r.Amount
		sumTurn += r.TurnoverRatio
		sumPrice += price
		cnt++
	}
	if cnt == 0 {
		return brief
	}
	brief.Price = sumPrice / cnt
	brief.ChangePct = sumPct / cnt
	brief.Amount = sumAmt
	brief.Turnover = sumTurn / cnt
	return brief
}

func (s *SinaSector) ListSectorStocks(ctx context.Context, sectorCode string, page, pageSize int) (*SectorStocksResult, error) {
	sectorCode = strings.TrimSpace(sectorCode)
	if sectorCode == "" {
		return nil, fmt.Errorf("invalid sector code")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	rows, err := s.fetchNodeData(ctx, sectorCode, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]StockBrief, 0, len(rows))
	for _, r := range rows {
		if r.Code == "" {
			continue
		}
		price, _ := strconv.ParseFloat(r.Trade, 64)
		items = append(items, StockBrief{
			Code:      r.Code,
			Name:      r.Name,
			Price:     price,
			ChangePct: r.ChangePercent,
		})
	}

	// 新浪接口不返回 total，用当前页数量估算
	total := len(items)
	if len(items) >= pageSize {
		total = page * pageSize
	}

	return &SectorStocksResult{
		SectorCode: sectorCode,
		Total:      total,
		Items:      items,
		Source:     "sina",
		UpdatedAt:  time.Now(),
	}, nil
}

func (s *SinaSector) fetchNodeData(ctx context.Context, node string, page, num int) ([]sinaHQRow, error) {
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("num", strconv.Itoa(num))
	params.Set("sort", "changepercent")
	params.Set("asc", "0")
	params.Set("node", node)

	u := sinaSectorBase + "/Market_Center.getHQNodeData?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if len(body) == 0 || string(body) == "null" {
		return nil, fmt.Errorf("sina: empty sector data for %s", node)
	}

	var rows []sinaHQRow
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}
	return rows, nil
}
