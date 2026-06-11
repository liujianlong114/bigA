package analysis

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

// Analyzer 持仓分析 + 大盘解析 + 异动检测 + 复盘报告
type Analyzer struct {
	db    *sql.DB
	redis *redis.Client
}

func New(db *sql.DB, rdb *redis.Client) *Analyzer {
	return &Analyzer{db: db, redis: rdb}
}

// liveQuote 从 Redis 实时行情 Hash 中解析的单只行情
type liveQuote struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	PrevClose float64 `json:"prev_close"`
	ChangePct float64 `json:"change_pct"`
	Volume    int64   `json:"volume"`
	Amount    float64 `json:"amount"`
	Turnover  float64 `json:"turnover"`
	Board     string  `json:"board"`
	LimitUp   float64 `json:"limit_up"`
	LimitDown float64 `json:"limit_down"`
}

func (a *Analyzer) getLiveQuotes(ctx context.Context) (map[string]*liveQuote, error) {
	raw, err := a.redis.HGetAll(ctx, "biga:live:quotes").Result()
	if err != nil {
		return nil, err
	}
	out := make(map[string]*liveQuote, len(raw))
	if len(raw) == 0 {
		return out, nil
	}
	for code, payload := range raw {
		var q liveQuote
		if err := json.Unmarshal([]byte(payload), &q); err != nil {
			continue
		}
		q.Code = code
		out[code] = &q
	}
	return out, nil
}

// getLiveQuote 获取单只实时行情
func (a *Analyzer) getLiveQuote(ctx context.Context, code string) (*liveQuote, error) {
	payload, err := a.redis.HGet(ctx, "biga:live:quotes", code).Bytes()
	if err != nil {
		return nil, err
	}
	var q liveQuote
	if err := json.Unmarshal(payload, &q); err != nil {
		return nil, err
	}
	q.Code = code
	return &q, nil
}

// getAvgVolume 从 K 线表获取近 N 日日均成交量
func (a *Analyzer) getAvgVolume(ctx context.Context, code string, days int) (float64, error) {
	var avg sql.NullFloat64
	err := a.db.QueryRowContext(ctx,
		`SELECT AVG(volume) FROM (
			SELECT volume FROM market_kline WHERE code=? AND period='day' ORDER BY trade_date DESC LIMIT ?
		) t`, code, days,
	).Scan(&avg)
	if err != nil || !avg.Valid {
		return 0, err
	}
	return avg.Float64, nil
}

// getClosePrices 获取最近N日收盘价
func (a *Analyzer) getClosePrices(ctx context.Context, code string, days int) ([]float64, error) {
	rows, err := a.db.QueryContext(ctx,
		`SELECT close_p FROM market_kline WHERE code=? AND period='day' ORDER BY trade_date DESC LIMIT ?`,
		code, days,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var prices []float64
	for rows.Next() {
		var p float64
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	// 反转成正序
	for i, j := 0, len(prices)-1; i < j; i, j = i+1, j-1 {
		prices[i], prices[j] = prices[j], prices[i]
	}
	return prices, rows.Err()
}

// calcReturns 计算日收益率序列
func calcReturns(prices []float64) []float64 {
	if len(prices) < 2 {
		return nil
	}
	ret := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1] != 0 {
			ret[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
		}
	}
	return ret
}

// calcBeta 计算 Beta 值
func calcBeta(stockReturns, indexReturns []float64) float64 {
	if len(stockReturns) == 0 || len(indexReturns) == 0 {
		return 1.0
	}
	n := len(stockReturns)
	if len(indexReturns) < n {
		n = len(indexReturns)
	}
	sr := stockReturns[len(stockReturns)-n:]
	ir := indexReturns[len(indexReturns)-n:]

	var sumSR, sumIR, sumSR2, sumSRIR float64
	for i := 0; i < n; i++ {
		sumSR += sr[i]
		sumIR += ir[i]
		sumSR2 += sr[i] * sr[i]
		sumSRIR += sr[i] * ir[i]
	}
	meanSR := sumSR / float64(n)
	meanIR := sumIR / float64(n)

	cov := sumSRIR/float64(n) - meanSR*meanIR
	vari := sumSR2/float64(n) - meanIR*meanIR
	if vari == 0 {
		return 1.0
	}
	return cov / vari
}

// calcMaxDrawdown 计算最大回撤
func calcMaxDrawdown(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}
	peak := prices[0]
	maxDD := 0.0
	for _, p := range prices {
		if p > peak {
			peak = p
		}
		dd := (peak - p) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

// calcMA 简单移动平均
func calcMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

// calcATR 计算 Average True Range
func calcATR(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		tr := math.Abs(prices[i] - prices[i-1])
		sum += tr
	}
	return sum / float64(period)
}

// calcStdDev 计算标准差
func calcStdDev(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))
	variance := 0.0
	for _, r := range returns {
		variance += (r - mean) * (r - mean)
	}
	return math.Sqrt(variance / float64(len(returns)))
}

// getIndustry 根据股票代码推断行业（基于申万一级行业映射，简化版）
func getIndustry(code string) string {
	// 使用代码前缀做粗略分类，实际应用应接入申万行业数据
	if len(code) < 3 {
		return "未知"
	}
	prefix := code[:3]
	switch {
	case prefix == "600" || prefix == "601" || prefix == "603" || prefix == "605":
		return inferIndustryByRange(code, "600000", "609999")
	case prefix == "000" || prefix == "001" || prefix == "002" || prefix == "003":
		return inferIndustryByRange(code, "000001", "004999")
	case prefix == "300" || prefix == "301":
		return "创业板"
	case prefix == "688":
		return "科创板"
	case prefix == "8" || prefix == "4" || prefix == "9":
		return "北交所"
	default:
		return "其他"
	}
}

func inferIndustryByRange(code, start, end string) string {
	// 实际工程中应接入申万行业分类 API 或本地映射表
	// 这里返回通用分类，待后续接入真实行业数据
	return "主板"
}

// sortByValueDesc 按 float 值降序返回排序后的 key-value pairs
type kvPair struct {
	Key   string
	Value float64
}

func sortByValueDesc(m map[string]float64) []kvPair {
	pairs := make([]kvPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kvPair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Value > pairs[j].Value })
	return pairs
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func formatFloat2(v float64) float64 {
	return math.Round(v*100) / 100
}

func formatFloat4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

func ptrStr(s string) *string { return &s }

func timeNow() string { return time.Now().Format(time.RFC3339) }

// 主要指数代码映射
var indexCodes = map[string]string{
	"000001": "上证指数",
	"399001": "深证成指",
	"399006": "创业板指",
	"000688": "科创50",
	"000300": "沪深300",
	"000905": "中证500",
	"000852": "中证1000",
}

// 为保证编译通过——避免 unused import
var _ = fmt.Sprintf
