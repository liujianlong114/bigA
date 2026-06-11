package analysis

import (
	"context"
	"fmt"
	"math"
)

// PredictionResult 涨跌推测结果
type PredictionResult struct {
	Code       string             `json:"code"`
	Name       string             `json:"name"`
	Score      int                `json:"score"`      // 0-100
	ProbUp1D   float64            `json:"prob_up_1d"` // 1日上涨概率
	ProbUp3D   float64            `json:"prob_up_3d"`
	ProbUp5D   float64            `json:"prob_up_5d"`
	Signal     string             `json:"signal"` // bullish / bearish / neutral
	Factors    map[string]float64 `json:"factors"`
	Support    float64            `json:"support"`
	Resistance float64            `json:"resistance"`
	Timestamp  string             `json:"timestamp"`
}

// Predict 多因子评分预测
func (a *Analyzer) Predict(ctx context.Context, code string) (*PredictionResult, error) {
	prices, err := a.getClosePrices(ctx, code, 120)
	if err != nil || len(prices) < 30 {
		return nil, fmt.Errorf("insufficient kline data for %s", code)
	}

	name := code
	if q, err := a.getLiveQuote(ctx, code); err == nil {
		name = q.Name
	}

	factors := map[string]float64{}
	score := 50.0

	// MA 趋势
	ma5 := calcMA(prices, 5)
	ma20 := calcMA(prices, 20)
	ma60 := calcMA(prices, 60)
	last := prices[len(prices)-1]
	if last > ma5 && ma5 > ma20 {
		factors["ma_trend"] = 15
		score += 15
	} else if last < ma5 && ma5 < ma20 {
		factors["ma_trend"] = -15
		score -= 15
	}
	if ma20 > ma60 {
		factors["ma_mid"] = 5
		score += 5
	} else {
		factors["ma_mid"] = -5
		score -= 5
	}

	// 动量：近5日 vs 前5日
	if len(prices) >= 10 {
		recent := (prices[len(prices)-1] - prices[len(prices)-6]) / prices[len(prices)-6]
		factors["momentum_5d"] = math.Round(recent*1000) / 10
		score += recent * 100
	}

	// RSI 14
	rsi := calcRSI(prices, 14)
	factors["rsi14"] = math.Round(rsi*10) / 10
	if rsi < 30 {
		score += 10
	} else if rsi > 70 {
		score -= 10
	}

	// 布林带位置
	upper, lower, mid := calcBOLL(prices, 20)
	if last <= lower {
		factors["boll"] = 10
		score += 10
	} else if last >= upper {
		factors["boll"] = -10
		score -= 10
	} else {
		factors["boll"] = (last - mid) / (upper - lower + 1e-9) * 10
		score += factors["boll"]
	}

	// 成交量
	avgVol, _ := a.getAvgVolume(ctx, code, 20)
	if q, err := a.getLiveQuote(ctx, code); err == nil && avgVol > 0 {
		vr := float64(q.Volume) / avgVol
		factors["vol_ratio"] = math.Round(vr*100) / 100
		if vr > 1.5 && q.ChangePct > 0 {
			score += 8
		} else if vr > 1.5 && q.ChangePct < 0 {
			score -= 8
		}
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	prob1 := score / 100
	prob3 := math.Min(1, prob1*0.9+0.05)
	prob5 := math.Min(1, prob1*0.85+0.075)

	signal := "neutral"
	if score >= 65 {
		signal = "bullish"
	} else if score <= 35 {
		signal = "bearish"
	}

	return &PredictionResult{
		Code:       code,
		Name:       name,
		Score:      int(math.Round(score)),
		ProbUp1D:   formatFloat4(prob1),
		ProbUp3D:   formatFloat4(prob3),
		ProbUp5D:   formatFloat4(prob5),
		Signal:     signal,
		Factors:    factors,
		Support:    formatFloat2(lower),
		Resistance: formatFloat2(upper),
		Timestamp:  timeNow(),
	}, nil
}

func calcRSI(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 50
	}
	gain, loss := 0.0, 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		d := prices[i] - prices[i-1]
		if d > 0 {
			gain += d
		} else {
			loss -= d
		}
	}
	if loss == 0 {
		return 100
	}
	rs := gain / loss
	return 100 - 100/(1+rs)
}

func calcBOLL(prices []float64, period int) (upper, lower, mid float64) {
	if len(prices) < period {
		return 0, 0, 0
	}
	mid = calcMA(prices, period)
	variance := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		d := prices[i] - mid
		variance += d * d
	}
	std := math.Sqrt(variance / float64(period))
	return mid + 2*std, mid - 2*std, mid
}
