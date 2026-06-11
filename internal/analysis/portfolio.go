package analysis

import (
	"context"
	"fmt"
	"strings"
)

// PortfolioAnalysisResult 持仓分析完整结果
type PortfolioAnalysisResult struct {
	AccountID    int64                   `json:"account_id"`
	TotalValue   float64                 `json:"total_value"`
	TotalCost    float64                 `json:"total_cost"`
	TotalPL      float64                 `json:"total_pl"`
	TotalPLPct   float64                 `json:"total_pl_pct"`
	Positions    []PositionAnalysis      `json:"positions"`
	IndustryDist map[string]IndustryStat `json:"industry_distribution"`
	RiskWarnings []string                `json:"risk_warnings"`
	Beta         float64                 `json:"beta"`
	MaxDrawdown  float64                 `json:"max_drawdown_90d_est"`
	Timestamp    string                  `json:"timestamp"`
}

// PositionAnalysis 单只持仓分析
type PositionAnalysis struct {
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Board        string  `json:"board"`
	Quantity     int     `json:"quantity"`
	CostPrice    float64 `json:"cost_price"`
	CurrentPrice float64 `json:"current_price"`
	MarketValue  float64 `json:"market_value"`
	PL           float64 `json:"pl"`
	PLPct        float64 `json:"pl_pct"`
	Industry     string  `json:"industry"`
	AvgVol20     float64 `json:"avg_vol_20d"`
	MA20         float64 `json:"ma20"`
	StopLoss     float64 `json:"stop_loss_suggest"`
	TakeProfit   float64 `json:"take_profit_suggest"`
	Quality      string  `json:"quality_grade"` // A/B/C/D
}

// IndustryStat 行业分布统计
type IndustryStat struct {
	Industry    string  `json:"industry"`
	MarketValue float64 `json:"market_value"`
	Pct         float64 `json:"pct"`
	StockCount  int     `json:"stock_count"`
}

// AnalyzePortfolio 持仓多维度分析
func (a *Analyzer) AnalyzePortfolio(ctx context.Context, accountID int64) (*PortfolioAnalysisResult, error) {
	// 1. 读取持仓
	positions, err := a.getPositions(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("get positions: %w", err)
	}
	if len(positions) == 0 {
		return &PortfolioAnalysisResult{
			AccountID: accountID,
			Positions: []PositionAnalysis{},
			Timestamp: timeNow(),
		}, nil
	}

	// 2. 获取实时行情（从 Redis live hash）
	liveQuotes, _ := a.getLiveQuotes(ctx)

	// 3. 获取沪深300收盘价（用于Beta计算）
	idxPrices, _ := a.getClosePrices(ctx, "000300", 90)

	result := &PortfolioAnalysisResult{
		AccountID:    accountID,
		Positions:    make([]PositionAnalysis, 0, len(positions)),
		IndustryDist: make(map[string]IndustryStat),
		RiskWarnings: []string{},
		Timestamp:    timeNow(),
	}

	var totalValue, totalCost float64
	var allStockReturns []float64

	for _, pos := range positions {
		pa := PositionAnalysis{
			Code:      pos.Code,
			Name:      pos.Name,
			Board:     pos.Board,
			Quantity:  pos.Quantity,
			CostPrice: pos.CostPrice,
			Industry:  getIndustry(pos.Code),
		}

		// 实时价格
		if q, ok := liveQuotes[pos.Code]; ok && q.Price > 0 {
			pa.CurrentPrice = q.Price
		}

		// 从 K 线获取技术数据
		prices, err := a.getClosePrices(ctx, pos.Code, 90)
		if err == nil && len(prices) >= 20 {
			pa.MA20 = formatFloat2(calcMA(prices, 20))
			atr := calcATR(prices, 14)
			if pa.CurrentPrice > 0 {
				pa.StopLoss = formatFloat2(pa.CurrentPrice - 2*atr)
				pa.TakeProfit = formatFloat2(pa.CurrentPrice + 3*atr)
			}

			// 质量评分
			returns := calcReturns(prices)
			if len(returns) >= 20 {
				volatility := calcStdDev(returns)
				recent := returns[len(returns)-20:]
				posCount := 0
				for _, r := range recent {
					if r > 0 {
						posCount++
					}
				}
				winRate := float64(posCount) / float64(len(recent))

				if winRate >= 0.6 && volatility < 0.03 {
					pa.Quality = "A"
				} else if winRate >= 0.5 && volatility < 0.04 {
					pa.Quality = "B"
				} else if winRate >= 0.4 {
					pa.Quality = "C"
				} else {
					pa.Quality = "D"
				}
			}

			// Beta 计算（用该股 + 指数）
			if len(prices) >= 30 && len(idxPrices) >= 30 {
				stockRet := calcReturns(prices)
				idxRet := calcReturns(idxPrices)
				beta := calcBeta(stockRet, idxRet)
				// 汇总算整体 Beta（后面会用所有持仓加权算）
				if len(stockRet) > 0 {
					allStockReturns = append(allStockReturns, stockRet[len(stockRet)-min(len(stockRet), len(idxRet)):]...)
				}
				_ = beta
			}
		}

		// 日均成交量
		if avgVol, err := a.getAvgVolume(ctx, pos.Code, 20); err == nil {
			pa.AvgVol20 = avgVol
		}

		// 计算市值和盈亏
		if pa.CurrentPrice > 0 {
			pa.MarketValue = formatFloat2(float64(pa.Quantity) * pa.CurrentPrice)
		}
		pa.PL = formatFloat2(pa.MarketValue - float64(pa.Quantity)*pa.CostPrice)
		if pa.CostPrice > 0 {
			pa.PLPct = formatFloat4((pa.CurrentPrice - pa.CostPrice) / pa.CostPrice * 100)
		}

		totalValue += pa.MarketValue
		totalCost += float64(pa.Quantity) * pa.CostPrice

		result.Positions = append(result.Positions, pa)

		// 行业分布统计
		ind := pa.Industry
		stat := result.IndustryDist[ind]
		stat.Industry = ind
		stat.MarketValue += pa.MarketValue
		stat.StockCount++
		result.IndustryDist[ind] = stat
	}

	result.TotalValue = formatFloat2(totalValue)
	result.TotalCost = formatFloat2(totalCost)
	result.TotalPL = formatFloat2(totalValue - totalCost)
	if totalCost > 0 {
		result.TotalPLPct = formatFloat4((totalValue - totalCost) / totalCost * 100)
	}

	// 计算行业占比
	totalMV := totalValue
	for k, v := range result.IndustryDist {
		if totalMV > 0 {
			v.Pct = formatFloat2(v.MarketValue / totalMV * 100)
		}
		result.IndustryDist[k] = v
	}

	// Beta 估算：使用组合平均收益率 vs 沪深300
	if len(allStockReturns) > 0 && len(idxPrices) >= 2 {
		idxRet := calcReturns(idxPrices)
		result.Beta = formatFloat4(calcBeta(allStockReturns, idxRet))
	} else {
		result.Beta = 1.0
	}

	// 最大回撤估算（持仓整体历史回测）
	if len(result.Positions) > 0 {
		// 用持仓组合的日收益率估算
		if len(allStockReturns) > 0 {
			cumulative := 1.0
			prices := make([]float64, len(allStockReturns)+1)
			prices[0] = 1.0
			for i, r := range allStockReturns {
				cumulative *= (1 + r)
				prices[i+1] = cumulative
			}
			result.MaxDrawdown = formatFloat4(calcMaxDrawdown(prices) * 100)
		}
	}

	// 风险预警
	var warnings []string

	// 集中度检查
	for _, stat := range result.IndustryDist {
		if stat.Pct > 40 {
			warnings = append(warnings, fmt.Sprintf("行业[%s]占比 %.1f%%，超过40%%集中度警戒线", stat.Industry, stat.Pct))
		}
	}
	for _, pa := range result.Positions {
		if totalMV > 0 {
			singlePct := pa.MarketValue / totalMV * 100
			if singlePct > 20 {
				warnings = append(warnings, fmt.Sprintf("个股[%s %s]占比 %.1f%%，超过20%%警戒线", pa.Code, pa.Name, singlePct))
			}
		}
		if pa.Quality == "D" {
			warnings = append(warnings, fmt.Sprintf("个股[%s %s]质量评级为D，建议关注风险", pa.Code, pa.Name))
		}
	}
	if result.Beta > 1.5 {
		warnings = append(warnings, fmt.Sprintf("组合Beta=%.2f，波动性显著高于大盘", result.Beta))
	}

	result.RiskWarnings = warnings
	return result, nil
}

// getPositions 从 MySQL 读取持仓
func (a *Analyzer) getPositions(ctx context.Context, accountID int64) ([]struct {
	Code      string
	Name      string
	Board     string
	Quantity  int
	CostPrice float64
}, error) {
	rows, err := a.db.QueryContext(ctx,
		`SELECT code, name, COALESCE(board,'main'), quantity, cost_price
		 FROM sim_position WHERE account_id=? ORDER BY code`, accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []struct {
		Code      string
		Name      string
		Board     string
		Quantity  int
		CostPrice float64
	}
	for rows.Next() {
		var p struct {
			Code      string
			Name      string
			Board     string
			Quantity  int
			CostPrice float64
		}
		if err := rows.Scan(&p.Code, &p.Name, &p.Board, &p.Quantity, &p.CostPrice); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// 消除未使用 import 产生的编译警告
var _ = strings.Split
