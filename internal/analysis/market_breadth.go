package analysis

import (
	"context"
	"fmt"
)

// MarketBreadthResult 大盘解析结果
type MarketBreadthResult struct {
	Indices       []IndexSnapshot  `json:"indices"`
	Breadth       BreadthStats     `json:"breadth"`
	Sentiment     SentimentStats   `json:"sentiment"`
	NorthBound    NorthBoundFlow   `json:"north_bound"`
	TopSectors    []SectorSnapshot `json:"top_sectors"`
	BottomSectors []SectorSnapshot `json:"bottom_sectors"`
	Timestamp     string           `json:"timestamp"`
}

// IndexSnapshot 指数快照
type IndexSnapshot struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	ChangePct float64 `json:"change_pct"`
	Volume    int64   `json:"volume"`
	Amount    float64 `json:"amount"`
}

// BreadthStats 市场宽度
type BreadthStats struct {
	Total       int     `json:"total_stocks"`
	Up          int     `json:"up"`
	Down        int     `json:"down"`
	Flat        int     `json:"flat"`
	UpDownRatio float64 `json:"up_down_ratio"`
	UpPct5Plus  int     `json:"up_pct_5plus"`
	DownPct5Min int     `json:"down_pct_5minus"`
}

// SentimentStats 情绪指标
type SentimentStats struct {
	LimitUp      int     `json:"limit_up_count"`
	LimitDown    int     `json:"limit_down_count"`
	AvgChangePct float64 `json:"avg_change_pct"`
	TotalVolume  int64   `json:"total_volume_est"`
	TotalAmount  float64 `json:"total_amount_est"`
}

// NorthBoundFlow 北向资金
type NorthBoundFlow struct {
	NetInflow float64 `json:"net_inflow"`
	Status    string  `json:"status"` // active / closed / no_data
}

// SectorSnapshot 板块快照
type SectorSnapshot struct {
	Name       string  `json:"name"`
	ChangePct  float64 `json:"change_pct"`
	StockCount int     `json:"stock_count"`
}

// AnalyzeMarketBreadth 大盘解析
func (a *Analyzer) AnalyzeMarketBreadth(ctx context.Context) (*MarketBreadthResult, error) {
	liveQuotes, err := a.getLiveQuotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get live quotes: %w", err)
	}

	result := &MarketBreadthResult{
		Indices:   make([]IndexSnapshot, 0),
		Timestamp: timeNow(),
	}

	// 1. 主要指数
	for code, name := range indexCodes {
		if q, ok := liveQuotes[code]; ok {
			result.Indices = append(result.Indices, IndexSnapshot{
				Code:      code,
				Name:      name,
				Price:     formatFloat2(q.Price),
				ChangePct: formatFloat4(q.ChangePct),
				Volume:    q.Volume,
				Amount:    formatFloat2(q.Amount),
			})
		}
	}

	// 2. 市场宽度统计
	var up, down, flat, up5, down5 int
	var sumChangePct, totalAmount float64
	var totalVolume int64

	for _, q := range liveQuotes {
		if q.Price <= 0 {
			continue
		}
		if q.ChangePct > 0.01 {
			up++
		} else if q.ChangePct < -0.01 {
			down++
		} else {
			flat++
		}
		if q.ChangePct >= 5 {
			up5++
		}
		if q.ChangePct <= -5 {
			down5++
		}
		sumChangePct += q.ChangePct
		totalVolume += q.Volume
		totalAmount += q.Amount
	}

	total := up + down + flat
	result.Breadth = BreadthStats{
		Total:       total,
		Up:          up,
		Down:        down,
		Flat:        flat,
		UpPct5Plus:  up5,
		DownPct5Min: down5,
	}
	if down > 0 {
		result.Breadth.UpDownRatio = formatFloat4(float64(up) / float64(down))
	} else if up > 0 {
		result.Breadth.UpDownRatio = 999
	}

	// 3. 情绪指标
	limitUp := 0
	limitDown := 0
	for _, q := range liveQuotes {
		if q.LimitUp > 0 && q.Price >= q.LimitUp*0.995 {
			limitUp++
		}
		if q.LimitDown > 0 && q.Price <= q.LimitDown*1.005 && q.Price > 0 {
			limitDown++
		}
	}
	result.Sentiment = SentimentStats{
		LimitUp:     limitUp,
		LimitDown:   limitDown,
		TotalVolume: totalVolume,
		TotalAmount: formatFloat2(totalAmount),
	}
	if total > 0 {
		result.Sentiment.AvgChangePct = formatFloat4(sumChangePct / float64(total))
	}

	// 4. 北向资金（从东方财富接口获取，暂用占位）
	result.NorthBound = NorthBoundFlow{
		NetInflow: 0,
		Status:    "no_data", // 后续接入东财沪股通/深股通接口
	}

	// 5. 板块排名（基于 industry 分组，从 live quotes 聚合）
	sectorMap := make(map[string]*SectorSnapshot)
	for _, q := range liveQuotes {
		ind := getIndustry(q.Code)
		if _, ok := sectorMap[ind]; !ok {
			sectorMap[ind] = &SectorSnapshot{Name: ind}
		}
		sectorMap[ind].StockCount++
		sectorMap[ind].ChangePct += q.ChangePct
	}
	for _, s := range sectorMap {
		if s.StockCount > 0 {
			s.ChangePct = formatFloat4(s.ChangePct / float64(s.StockCount))
		}
	}

	// 排序取前5后5
	var sectors []SectorSnapshot
	for _, s := range sectorMap {
		sectors = append(sectors, *s)
	}
	// bubble sort by ChangePct desc
	for i := 0; i < len(sectors); i++ {
		for j := i + 1; j < len(sectors); j++ {
			if sectors[j].ChangePct > sectors[i].ChangePct {
				sectors[i], sectors[j] = sectors[j], sectors[i]
			}
		}
	}
	if len(sectors) > 5 {
		result.TopSectors = sectors[:5]
	} else {
		result.TopSectors = sectors
	}
	if len(sectors) > 5 {
		last := sectors[len(sectors)-5:]
		// 反转
		for i, j := 0, len(last)-1; i < j; i, j = i+1, j-1 {
			last[i], last[j] = last[j], last[i]
		}
		result.BottomSectors = last
	} else if len(sectors) > 0 {
		rev := make([]SectorSnapshot, len(sectors))
		copy(rev, sectors)
		for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
			rev[i], rev[j] = rev[j], rev[i]
		}
		result.BottomSectors = rev
	}

	return result, nil
}

var _ = fmt.Sprintf
