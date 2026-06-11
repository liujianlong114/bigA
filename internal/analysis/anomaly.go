package analysis

import (
	"context"
	"fmt"
)

// AnomalyResult 异动检测结果
type AnomalyResult struct {
	Anomalies []AnomalyItem `json:"anomalies"`
	Count     int           `json:"count"`
	Timestamp string        `json:"timestamp"`
}

// AnomalyItem 单条异动
type AnomalyItem struct {
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	Price     float64  `json:"price"`
	ChangePct float64  `json:"change_pct"`
	Volume    int64    `json:"volume"`
	Turnover  float64  `json:"turnover"`
	AvgVol20  float64  `json:"avg_vol_20d"`
	VolRatio  float64  `json:"vol_ratio"`
	Types     []string `json:"anomaly_types"`
	Causes    []string `json:"possible_causes"`
	Board     string   `json:"board"`
}

// DetectAnomalies 异动检测
func (a *Analyzer) DetectAnomalies(ctx context.Context) (*AnomalyResult, error) {
	liveQuotes, err := a.getLiveQuotes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get live quotes: %w", err)
	}

	result := &AnomalyResult{
		Anomalies: make([]AnomalyItem, 0),
		Timestamp: timeNow(),
	}

	for _, q := range liveQuotes {
		if q.Price <= 0 {
			continue
		}

		var types []string

		// 涨跌幅异常
		absChg := q.ChangePct
		if absChg < 0 {
			absChg = -absChg
		}
		switch q.Board {
		case "chinext", "star":
			if absChg >= 15 {
				types = append(types, "涨跌幅异常(双创≥15%)")
			} else if absChg >= 7 {
				types = append(types, "涨跌幅异常(≥7%)")
			}
		default:
			if absChg >= 7 {
				types = append(types, "涨跌幅异常(≥7%)")
			}
		}

		// 成交量异常（需要对比20日均量）
		avgVol, err := a.getAvgVolume(ctx, q.Code, 20)
		volRatio := 0.0
		if err == nil && avgVol > 0 {
			volRatio = float64(q.Volume) / avgVol
			if volRatio >= 3 {
				types = append(types, fmt.Sprintf("放量%.1f倍(相对20日均量)", volRatio))
			}
		}

		// 换手率异常
		if q.Turnover > 20 {
			types = append(types, fmt.Sprintf("换手率异常(%.1f%%)", q.Turnover))
		} else if q.Turnover > 10 {
			types = append(types, fmt.Sprintf("换手率偏高(%.1f%%)", q.Turnover))
		}

		if len(types) == 0 {
			continue
		}

		// 推测原因
		causes := inferCauses(q, volRatio)

		result.Anomalies = append(result.Anomalies, AnomalyItem{
			Code:      q.Code,
			Name:      q.Name,
			Price:     formatFloat2(q.Price),
			ChangePct: formatFloat4(q.ChangePct),
			Volume:    q.Volume,
			Turnover:  formatFloat2(q.Turnover),
			AvgVol20:  formatFloat2(avgVol),
			VolRatio:  formatFloat4(volRatio),
			Types:     types,
			Causes:    causes,
			Board:     q.Board,
		})
	}

	result.Count = len(result.Anomalies)
	return result, nil
}

// inferCauses 根据行情特征推测可能原因
func inferCauses(q *liveQuote, volRatio float64) []string {
	var causes []string

	if q.ChangePct > 5 && volRatio > 2 {
		causes = append(causes, "放量上涨，可能有重大利好消息或资金介入")
	}
	if q.ChangePct < -5 && volRatio > 2 {
		causes = append(causes, "放量下跌，可能有利空消息或主力出货")
	}
	if q.ChangePct > 9.5 && q.Board != "chinext" && q.Board != "star" {
		causes = append(causes, "接近涨停，关注封板力度和连板可能")
	}
	if q.ChangePct < -9.5 && q.Board != "chinext" && q.Board != "star" {
		causes = append(causes, "接近跌停，注意风险，可能存在基本面利空")
	}
	if q.Turnover > 20 {
		causes = append(causes, "超高换手率，可能为新股上市或主力对倒")
	}
	if volRatio > 5 {
		causes = append(causes, "极端放量，关注是否有重大公告或龙虎榜异动")
	}
	if len(causes) == 0 {
		causes = append(causes, "异动原因待确认，建议关注近期公告和新闻")
	}

	return causes
}

var _ = fmt.Sprintf
