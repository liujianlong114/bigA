package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// KlinePeriod K 线周期
type KlinePeriod string

const (
	PeriodDay   KlinePeriod = "day"
	PeriodWeek  KlinePeriod = "week"
	PeriodMonth KlinePeriod = "month"
	PeriodMin5  KlinePeriod = "m5"
	PeriodMin15 KlinePeriod = "m15"
	PeriodMin30 KlinePeriod = "m30"
	PeriodMin60 KlinePeriod = "m60"
)

// KlineBar 单根 K 线
type KlineBar struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
	Amount float64 `json:"amount"`
}

// KlineResult K 线查询结果
type KlineResult struct {
	Code   string      `json:"code"`
	Name   string      `json:"name"`
	Period KlinePeriod `json:"period"`
	Source string      `json:"source"`
	Bars   []KlineBar  `json:"bars"`
}

func (p KlinePeriod) eastmoneyKLT() string {
	switch p {
	case PeriodWeek:
		return "102"
	case PeriodMonth:
		return "103"
	case PeriodMin5:
		return "5"
	case PeriodMin15:
		return "15"
	case PeriodMin30:
		return "30"
	case PeriodMin60:
		return "60"
	default:
		return "101"
	}
}

func (p KlinePeriod) sinaScale() string {
	switch p {
	case PeriodMin5:
		return "5"
	case PeriodMin15:
		return "15"
	case PeriodMin30:
		return "30"
	case PeriodMin60:
		return "60"
	case PeriodWeek:
		return "1200"
	case PeriodMonth:
		return "7200"
	default:
		return "240"
	}
}

// KlineFetcher K 线数据源
type KlineFetcher struct {
	client *http.Client
	sina   *Sina
}

func NewKlineFetcher() *KlineFetcher {
	return &KlineFetcher{
		client: &http.Client{Timeout: 20 * time.Second},
		sina:   NewSina(),
	}
}

// Fetch 拉取 K 线（东方财富 push2his 优先，新浪兜底）
func (k *KlineFetcher) Fetch(ctx context.Context, code string, period KlinePeriod, limit int) (*KlineResult, error) {
	if limit < 1 || limit > 500 {
		limit = 120
	}
	res, err := k.fetchEastMoney(ctx, code, period, limit)
	if err == nil && len(res.Bars) > 0 {
		return res, nil
	}
	return k.fetchSina(ctx, code, period, limit)
}

func (k *KlineFetcher) fetchEastMoney(ctx context.Context, code string, period KlinePeriod, limit int) (*KlineResult, error) {
	secid := SecID(code)
	if secid == "" {
		return nil, fmt.Errorf("invalid code")
	}
	params := url.Values{}
	params.Set("secid", secid)
	params.Set("fields1", "f1,f2,f3,f4,f5,f6")
	params.Set("fields2", "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61")
	params.Set("klt", period.eastmoneyKLT())
	params.Set("fqt", "1")
	params.Set("end", "20500101")
	params.Set("lmt", strconv.Itoa(limit))

	path := "/api/qt/stock/kline/get?" + params.Encode()
	urls := eastmoneyHosts(func(host string) string { return host + path })

	body, err := resilientGet(ctx, k.client, urls, defaultHeaders())
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data struct {
			Code   string   `json:"code"`
			Name   string   `json:"name"`
			Klines []string `json:"klines"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	bars := parseEastMoneyKlines(resp.Data.Klines)
	return &KlineResult{
		Code: code, Name: resp.Data.Name, Period: period, Source: "eastmoney", Bars: bars,
	}, nil
}

func parseEastMoneyKlines(lines []string) []KlineBar {
	out := make([]KlineBar, 0, len(lines))
	for _, line := range lines {
		p := strings.Split(line, ",")
		if len(p) < 7 {
			continue
		}
		out = append(out, KlineBar{
			Date:   p[0],
			Open:   parseF(p[1]),
			Close:  parseF(p[2]),
			High:   parseF(p[3]),
			Low:    parseF(p[4]),
			Volume: int64(parseF(p[5])),
			Amount: parseF(p[6]),
		})
	}
	return out
}

func (k *KlineFetcher) fetchSina(ctx context.Context, code string, period KlinePeriod, limit int) (*KlineResult, error) {
	sym := SinaSymbol(code)
	if sym == "" {
		return nil, fmt.Errorf("invalid code")
	}
	raw := fmt.Sprintf(
		"https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=%s&ma=no&datalen=%d",
		sym, period.sinaScale(), limit,
	)
	sl := NewSinaList()
	var rows []struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	if err := sl.fetchJSON(ctx, raw, &rows); err != nil {
		return nil, err
	}
	name := code
	if q, err := k.sina.Quote(ctx, code); err == nil {
		name = q.Name
	}
	bars := make([]KlineBar, 0, len(rows))
	for _, r := range rows {
		bars = append(bars, KlineBar{
			Date:   r.Day,
			Open:   parseF(r.Open),
			High:   parseF(r.High),
			Low:    parseF(r.Low),
			Close:  parseF(r.Close),
			Volume: int64(parseF(r.Volume)),
		})
	}
	return &KlineResult{Code: code, Name: name, Period: period, Source: "sina", Bars: bars}, nil
}

func parseF(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
