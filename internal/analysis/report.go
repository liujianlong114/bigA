package analysis

import (
	"context"
	"fmt"
	"time"
)

// DailyReport 每日复盘报告
type DailyReport struct {
	Date            string             `json:"date"`
	MarketSummary   MarketSummary      `json:"market_summary"`
	SectorRank      SectorRanking      `json:"sector_ranking"`
	TopMovers       []MoverItem        `json:"top_movers"`
	BottomMovers    []MoverItem        `json:"bottom_movers"`
	Anomalies       []AnomalyItem      `json:"anomalies"`
	Watchlist       []WatchItem        `json:"tomorrow_watchlist"`
	PortfolioReview *PortfolioSnapshot `json:"portfolio_review,omitempty"`
	GeneratedAt     string             `json:"generated_at"`
}

// MarketSummary 大盘综述
type MarketSummary struct {
	Status       string  `json:"status"` // 强势/震荡/弱势
	Description  string  `json:"description"`
	UpDownRatio  float64 `json:"up_down_ratio"`
	LimitUpCount int     `json:"limit_up_count"`
	AvgChangePct float64 `json:"avg_change_pct"`
	TotalAmount  float64 `json:"total_amount"`
}

// SectorRanking 板块排名
type SectorRanking struct {
	Top3    []SectorSnapshot `json:"top3"`
	Bottom3 []SectorSnapshot `json:"bottom3"`
}

// MoverItem 涨跌榜项目
type MoverItem struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	ChangePct float64 `json:"change_pct"`
	Volume    int64   `json:"volume"`
}

// WatchItem 明日关注
type WatchItem struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// PortfolioSnapshot 持仓快照
type PortfolioSnapshot struct {
	TotalValue float64 `json:"total_value"`
	TotalPL    float64 `json:"total_pl"`
	TotalPLPct float64 `json:"total_pl_pct"`
	PosCount   int     `json:"position_count"`
}

// GenerateDailyReport 生成每日复盘报告
func (a *Analyzer) GenerateDailyReport(ctx context.Context, accountID int64) (*DailyReport, error) {
	today := time.Now().Format("2006-01-02")

	report := &DailyReport{
		Date:        today,
		GeneratedAt: timeNow(),
	}

	// 1. 大盘数据
	breadth, err := a.AnalyzeMarketBreadth(ctx)
	if err != nil {
		return nil, fmt.Errorf("market breadth: %w", err)
	}

	// 大盘综述
	report.MarketSummary = MarketSummary{
		UpDownRatio:  breadth.Breadth.UpDownRatio,
		LimitUpCount: breadth.Sentiment.LimitUp,
		AvgChangePct: breadth.Sentiment.AvgChangePct,
		TotalAmount:  breadth.Sentiment.TotalAmount,
	}

	if breadth.Breadth.UpDownRatio > 2 && breadth.Sentiment.AvgChangePct > 1 {
		report.MarketSummary.Status = "强势"
		report.MarketSummary.Description = "市场普涨，赚钱效应显著"
	} else if breadth.Breadth.UpDownRatio > 1.2 && breadth.Sentiment.AvgChangePct > 0 {
		report.MarketSummary.Status = "偏强震荡"
		report.MarketSummary.Description = "涨多跌少，结构性行情"
	} else if breadth.Breadth.UpDownRatio < 0.5 {
		report.MarketSummary.Status = "弱势"
		report.MarketSummary.Description = "市场普遍下跌，注意风险控制"
	} else if breadth.Sentiment.AvgChangePct < -0.5 {
		report.MarketSummary.Status = "偏弱震荡"
		report.MarketSummary.Description = "跌多涨少，观望为主"
	} else {
		report.MarketSummary.Status = "震荡"
		report.MarketSummary.Description = "多空力量均衡，个股分化"
	}

	// 板块排名
	if len(breadth.TopSectors) >= 3 {
		report.SectorRank.Top3 = breadth.TopSectors[:3]
	} else {
		report.SectorRank.Top3 = breadth.TopSectors
	}
	if len(breadth.BottomSectors) >= 3 {
		report.SectorRank.Bottom3 = breadth.BottomSectors[:3]
	} else {
		report.SectorRank.Bottom3 = breadth.BottomSectors
	}

	// 2. 异动检测
	anomalyRes, _ := a.DetectAnomalies(ctx)
	if anomalyRes != nil {
		report.Anomalies = anomalyRes.Anomalies
	}

	// 3. 涨跌榜（从 live quotes 取 Top5/Bottom5）
	liveQuotes, _ := a.getLiveQuotes(ctx)
	if liveQuotes != nil {
		var movers []MoverItem
		for _, q := range liveQuotes {
			if q.Price <= 0 {
				continue
			}
			movers = append(movers, MoverItem{
				Code: q.Code, Name: q.Name, Price: q.Price,
				ChangePct: q.ChangePct, Volume: q.Volume,
			})
		}
		// 按 ChangePct 排序
		type moverPair struct {
			item MoverItem
			pct  float64
		}
		var pairs []moverPair
		for _, m := range movers {
			pairs = append(pairs, moverPair{m, m.ChangePct})
		}
		// 简单选择排序
		for i := 0; i < len(pairs); i++ {
			for j := i + 1; j < len(pairs); j++ {
				if pairs[j].pct > pairs[i].pct {
					pairs[i], pairs[j] = pairs[j], pairs[i]
				}
			}
		}
		topN := 5
		if len(pairs) < topN {
			topN = len(pairs)
		}
		report.TopMovers = make([]MoverItem, topN)
		for i := 0; i < topN; i++ {
			report.TopMovers[i] = pairs[i].item
		}
		// 后5
		bottomN := 5
		if len(pairs) < bottomN {
			bottomN = len(pairs)
		}
		report.BottomMovers = make([]MoverItem, bottomN)
		for i := 0; i < bottomN; i++ {
			report.BottomMovers[i] = pairs[len(pairs)-1-i].item
		}
	}

	// 4. 明日关注（连续放量 + 异动股）
	seen := make(map[string]bool)
	if anomalyRes != nil {
		for _, an := range anomalyRes.Anomalies {
			if seen[an.Code] {
				continue
			}
			seen[an.Code] = true
			reason := ""
			if len(an.Types) > 0 {
				reason = an.Types[0]
			}
			report.Watchlist = append(report.Watchlist, WatchItem{
				Code: an.Code, Name: an.Name, Reason: reason,
			})
		}
	}

	// 5. 持仓回顾
	if accountID > 0 {
		portRes, err := a.AnalyzePortfolio(ctx, accountID)
		if err == nil {
			report.PortfolioReview = &PortfolioSnapshot{
				TotalValue: portRes.TotalValue,
				TotalPL:    portRes.TotalPL,
				TotalPLPct: portRes.TotalPLPct,
				PosCount:   len(portRes.Positions),
			}
		}
	}

	return report, nil
}

var _ = fmt.Sprintf
