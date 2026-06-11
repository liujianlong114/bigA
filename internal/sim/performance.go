package sim

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/store"
)

const benchmarkCode = "000300" // 沪深300

// DailyNavPoint 每日净值
type DailyNavPoint struct {
	Date           string  `json:"date"`
	Nav            float64 `json:"nav"`
	DailyReturnPct float64 `json:"daily_return_pct"`
	CumulativePct  float64 `json:"cumulative_return_pct"`
}

// PerformanceReport 账户绩效
type PerformanceReport struct {
	AccountID          int64           `json:"account_id"`
	InitialCash        float64         `json:"initial_cash"`
	CurrentNav         float64         `json:"current_nav"`
	TotalReturn        float64         `json:"total_return"`
	TotalReturnPct     float64         `json:"total_return_pct"`
	Benchmark          string          `json:"benchmark"`
	BenchmarkReturnPct float64         `json:"benchmark_return_pct"`
	ExcessReturnPct    float64         `json:"excess_return_pct"`
	MaxDrawdownPct     float64         `json:"max_drawdown_pct"`
	SharpeRatio        float64         `json:"sharpe_ratio"`
	CalmarRatio        float64         `json:"calmar_ratio"`
	WinRate            float64         `json:"win_rate"`
	ProfitLossRatio    float64         `json:"profit_loss_ratio"`
	AvgHoldingDays     float64         `json:"avg_holding_days"`
	ClosedTrades       int             `json:"closed_trades"`
	DailyNav           []DailyNavPoint `json:"daily_nav"`
	BenchmarkNav       []DailyNavPoint `json:"benchmark_nav"`
	UpdatedAt          string          `json:"updated_at"`
}

// Performance 绩效分析
type Performance struct {
	repo        *store.Repo
	market      *market.Service
	portfolio   *Portfolio
	initialCash float64
}

func NewPerformance(repo *store.Repo, ms *market.Service, pf *Portfolio, initialCash float64) *Performance {
	return &Performance{repo: repo, market: ms, portfolio: pf, initialCash: initialCash}
}

func (p *Performance) Analyze(ctx context.Context, accountID int64) (*PerformanceReport, error) {
	summary, err := p.portfolio.Summary(ctx, accountID)
	if err != nil {
		return nil, err
	}
	currentNav := summary.TotalAssets

	trades, err := p.repo.ListTradesAsc(ctx, accountID)
	if err != nil {
		return nil, err
	}

	initial := p.initialCash
	if initial <= 0 {
		initial = currentNav
	}

	report := &PerformanceReport{
		AccountID:    accountID,
		InitialCash:  initial,
		CurrentNav:   round2(currentNav),
		Benchmark:    benchmarkCode,
		UpdatedAt:    time.Now().Format(time.RFC3339),
		BenchmarkNav: []DailyNavPoint{},
	}

	winRate, plRatio, avgHold, closed := analyzeRoundTrips(trades)
	report.WinRate = winRate
	report.ProfitLossRatio = plRatio
	report.AvgHoldingDays = avgHold
	report.ClosedTrades = closed

	dailyNav := buildDailyNav(ctx, p, initial, trades, currentNav)
	report.DailyNav = dailyNav

	report.TotalReturn = round2(currentNav - initial)
	if initial > 0 {
		report.TotalReturnPct = round2((currentNav/initial - 1) * 100)
	}

	if len(dailyNav) > 0 {
		report.MaxDrawdownPct = round2(calcMaxDrawdownPct(dailyNav) * 100)
		returns := dailyReturns(dailyNav)
		report.SharpeRatio = round4(calcSharpe(returns))
		report.CalmarRatio = round4(calcCalmar(report.TotalReturnPct, report.MaxDrawdownPct, len(dailyNav)))
	}

	benchNav, benchRet := buildBenchmarkNav(ctx, p.market, dailyNav)
	report.BenchmarkNav = benchNav
	report.BenchmarkReturnPct = round2(benchRet)
	report.ExcessReturnPct = round2(report.TotalReturnPct - benchRet)

	return report, nil
}

type lot struct {
	qty     int
	cost    float64
	buyDate time.Time
	buyFees float64
}

func analyzeRoundTrips(trades []store.Trade) (winRate, plRatio, avgHold float64, closed int) {
	lots := map[string][]lot{}
	var pnls []float64
	var holds []float64

	for _, t := range trades {
		if t.Side == "buy" {
			fees := t.Commission + t.TransferFee
			lots[t.Code] = append(lots[t.Code], lot{
				qty: t.Quantity, cost: t.Amount, buyDate: t.TradedAt, buyFees: fees,
			})
			continue
		}
		remain := t.Quantity
		sellFees := t.Commission + t.StampTax + t.TransferFee
		sellProceeds := t.Amount
		for remain > 0 && len(lots[t.Code]) > 0 {
			head := lots[t.Code][0]
			take := remain
			if take > head.qty {
				take = head.qty
			}
			ratio := float64(take) / float64(t.Quantity)
			costPart := head.cost * float64(take) / float64(head.qty)
			buyFeePart := head.buyFees * float64(take) / float64(head.qty)
			feePart := buyFeePart + sellFees*ratio
			proceedsPart := sellProceeds * ratio
			pnl := proceedsPart - costPart - feePart
			pnls = append(pnls, pnl)
			holds = append(holds, t.TradedAt.Sub(head.buyDate).Hours()/24)

			head.qty -= take
			head.cost -= costPart
			head.buyFees -= buyFeePart
			if head.qty == 0 {
				lots[t.Code] = lots[t.Code][1:]
			} else {
				lots[t.Code][0] = head
			}
			remain -= take
		}
	}

	closed = len(pnls)
	if closed == 0 {
		return 0, 0, 0, 0
	}

	wins := 0
	var winSum, lossSum float64
	for _, pnl := range pnls {
		if pnl > 0 {
			wins++
			winSum += pnl
		} else if pnl < 0 {
			lossSum += math.Abs(pnl)
		}
	}
	winRate = round2(float64(wins) / float64(closed) * 100)
	if lossSum > 0 {
		plRatio = round2(winSum / lossSum)
	} else if winSum > 0 {
		plRatio = 999
	}
	var holdSum float64
	for _, h := range holds {
		holdSum += h
	}
	avgHold = round2(holdSum / float64(len(holds)))
	return winRate, plRatio, avgHold, closed
}

func buildDailyNav(ctx context.Context, p *Performance, initial float64, trades []store.Trade, currentNav float64) []DailyNavPoint {
	today := time.Now().Format("2006-01-02")
	if len(trades) == 0 {
		return []DailyNavPoint{{
			Date: today, Nav: round2(currentNav), DailyReturnPct: 0,
			CumulativePct: round2((currentNav/initial - 1) * 100),
		}}
	}

	cash := initial
	positions := map[string]int{}
	closeCache := map[string]map[string]float64{} // code -> date -> close

	getClose := func(code, date string) float64 {
		if byDate, ok := closeCache[code]; ok {
			if c, ok := byDate[date]; ok {
				return c
			}
		}
		if closeCache[code] == nil {
			closeCache[code] = map[string]float64{}
		}
		bars, err := p.market.KlineFromDB(ctx, code, market.PeriodDay, 250)
		if err != nil || len(bars) == 0 {
			res, err := p.market.GetKline(ctx, code, market.PeriodDay, 250)
			if err == nil {
				for _, b := range res.Bars {
					closeCache[code][b.Date] = b.Close
				}
			}
		} else {
			for _, b := range bars {
				closeCache[code][b.Date] = b.Close
			}
		}
		if c, ok := closeCache[code][date]; ok {
			return c
		}
		// fallback: 最近不晚于 date 的收盘价
		var best string
		for d := range closeCache[code] {
			if d <= date && (best == "" || d > best) {
				best = d
			}
		}
		if best != "" {
			return closeCache[code][best]
		}
		return 0
	}

	valuate := func(date string) float64 {
		nav := cash
		for code, qty := range positions {
			if qty <= 0 {
				continue
			}
			px := getClose(code, date)
			if px <= 0 {
				continue
			}
			nav += px * float64(qty)
		}
		return round2(nav)
	}

	// 按日分组交易
	byDate := map[string][]store.Trade{}
	dates := map[string]struct{}{}
	for _, t := range trades {
		d := t.TradeDate
		byDate[d] = append(byDate[d], t)
		dates[d] = struct{}{}
	}
	dates[today] = struct{}{}

	sorted := make([]string, 0, len(dates))
	for d := range dates {
		sorted = append(sorted, d)
	}
	sort.Strings(sorted)

	var points []DailyNavPoint
	prevNav := initial
	for _, d := range sorted {
		for _, t := range byDate[d] {
			fees := t.Commission + t.TransferFee
			if t.Side == "sell" {
				fees += t.StampTax
			}
			if t.Side == "buy" {
				cash -= t.Amount + fees
				positions[t.Code] += t.Quantity
			} else {
				cash += t.Amount - fees
				positions[t.Code] -= t.Quantity
				if positions[t.Code] <= 0 {
					delete(positions, t.Code)
				}
			}
		}
		nav := valuate(d)
		if d == today {
			nav = round2(currentNav)
		}
		dailyRet := 0.0
		if prevNav > 0 {
			dailyRet = round4((nav/prevNav - 1) * 100)
		}
		cumRet := 0.0
		if initial > 0 {
			cumRet = round2((nav/initial - 1) * 100)
		}
		points = append(points, DailyNavPoint{
			Date: d, Nav: nav, DailyReturnPct: dailyRet, CumulativePct: cumRet,
		})
		prevNav = nav
	}
	return points
}

func buildBenchmarkNav(ctx context.Context, ms *market.Service, dailyNav []DailyNavPoint) ([]DailyNavPoint, float64) {
	if len(dailyNav) == 0 {
		return nil, 0
	}
	startDate := dailyNav[0].Date
	endDate := dailyNav[len(dailyNav)-1].Date

	bars, err := ms.KlineFromDB(ctx, benchmarkCode, market.PeriodDay, 300)
	if err != nil || len(bars) == 0 {
		res, err := ms.GetKline(ctx, benchmarkCode, market.PeriodDay, 300)
		if err != nil || len(res.Bars) == 0 {
			return []DailyNavPoint{}, 0
		}
		bars = res.Bars
	}

	closeByDate := map[string]float64{}
	for _, b := range bars {
		closeByDate[b.Date] = b.Close
	}

	var startClose, endClose float64
	for _, b := range bars {
		if b.Date <= startDate {
			startClose = b.Close
		}
		if b.Date <= endDate {
			endClose = b.Close
		}
	}
	if startClose <= 0 || endClose <= 0 {
		return []DailyNavPoint{}, 0
	}

	base := 100.0
	var out []DailyNavPoint
	prev := base
	for _, pt := range dailyNav {
		c := closeByDate[pt.Date]
		if c <= 0 {
			for _, b := range bars {
				if b.Date <= pt.Date {
					c = b.Close
				}
			}
		}
		if c <= 0 {
			continue
		}
		nav := round2(base * c / startClose)
		dailyRet := 0.0
		if prev > 0 {
			dailyRet = round4((nav/prev - 1) * 100)
		}
		cumRet := round2((nav/base - 1) * 100)
		out = append(out, DailyNavPoint{
			Date: pt.Date, Nav: nav, DailyReturnPct: dailyRet, CumulativePct: cumRet,
		})
		prev = nav
	}

	benchRet := round2((endClose/startClose - 1) * 100)
	return out, benchRet
}

func dailyReturns(nav []DailyNavPoint) []float64 {
	if len(nav) < 2 {
		return nil
	}
	ret := make([]float64, len(nav)-1)
	for i := 1; i < len(nav); i++ {
		if nav[i-1].Nav > 0 {
			ret[i-1] = (nav[i].Nav - nav[i-1].Nav) / nav[i-1].Nav
		}
	}
	return ret
}

func calcMaxDrawdownPct(nav []DailyNavPoint) float64 {
	if len(nav) == 0 {
		return 0
	}
	peak := nav[0].Nav
	maxDD := 0.0
	for _, p := range nav {
		if p.Nav > peak {
			peak = p.Nav
		}
		if peak > 0 {
			dd := (peak - p.Nav) / peak
			if dd > maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD
}

func calcSharpe(dailyReturns []float64) float64 {
	if len(dailyReturns) < 2 {
		return 0
	}
	mean := 0.0
	for _, r := range dailyReturns {
		mean += r
	}
	mean /= float64(len(dailyReturns))
	var varSum float64
	for _, r := range dailyReturns {
		varSum += (r - mean) * (r - mean)
	}
	std := math.Sqrt(varSum / float64(len(dailyReturns)))
	if std == 0 {
		return 0
	}
	// 年化夏普（无风险利率简化为 0）
	return mean / std * math.Sqrt(252)
}

func calcCalmar(totalReturnPct, maxDrawdownPct float64, days int) float64 {
	if maxDrawdownPct <= 0 || days <= 0 {
		return 0
	}
	annualReturn := totalReturnPct * 365 / float64(days)
	return annualReturn / maxDrawdownPct
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}
