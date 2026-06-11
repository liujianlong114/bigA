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

const (
	eastMoneyQuotePath = "/api/qt/stock/get"
	eastMoneyListPath  = "/api/qt/clist/get"
	eastMoneyUT        = "fa5fd1943c7b386f172d6893dbfba10b"
)

// EastMoney 东方财富免费行情
type EastMoney struct {
	client *http.Client
}

func NewEastMoney() *EastMoney {
	return &EastMoney{
		client: &http.Client{Timeout: 12 * time.Second},
	}
}

func (e *EastMoney) doGet(ctx context.Context, apiPath string, params url.Values) ([]byte, error) {
	urls := eastmoneyHosts(func(host string) string {
		return host + apiPath + "?" + params.Encode()
	})
	return resilientGet(ctx, e.client, urls, defaultHeaders())
}

// Quote 获取单只股票实时行情
func (e *EastMoney) Quote(ctx context.Context, code string) (*Quote, error) {
	secid := SecID(code)
	if secid == "" {
		return nil, fmt.Errorf("invalid stock code: %s", code)
	}
	params := url.Values{}
	params.Set("ut", eastMoneyUT)
	params.Set("invt", "2")
	params.Set("fltt", "2")
	params.Set("secid", secid)
	params.Set("fields", "f57,f58,f43,f44,f45,f46,f47,f48,f60,f169,f170,f19,f20")

	body, err := e.doGet(ctx, eastMoneyQuotePath, params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, fmt.Errorf("eastmoney: empty data for %s", code)
	}

	f := func(key string) float64 {
		v, ok := resp.Data[key]
		if !ok {
			return 0
		}
		var n float64
		if err := json.Unmarshal(v, &n); err != nil {
			return 0
		}
		return n
	}
	name := ""
	if raw, ok := resp.Data["f58"]; ok {
		_ = json.Unmarshal(raw, &name)
	}
	sym := ""
	if raw, ok := resp.Data["f57"]; ok {
		_ = json.Unmarshal(raw, &sym)
	}
	if sym == "" {
		sym = code
	}

	price := f("f43")
	prev := f("f60")
	change := f("f169")
	if change == 0 && prev > 0 {
		change = price - prev
	}
	pct := f("f170")

	return &Quote{
		Code:      sym,
		Name:      name,
		Price:     price,
		Open:      f("f46"),
		High:      f("f44"),
		Low:       f("f45"),
		PrevClose: prev,
		Change:    change,
		ChangePct: pct,
		Volume:    int64(f("f47")),
		Amount:    f("f48"),
		Bid1:      f("f19"),
		Ask1:      f("f20"),
		Source:    "eastmoney",
		UpdatedAt: time.Now(),
	}, nil
}

// List 沪深 A 股列表（分页）
func (e *EastMoney) List(ctx context.Context, page, pageSize int) ([]StockBrief, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	params := url.Values{}
	params.Set("pn", strconv.Itoa(page))
	params.Set("pz", strconv.Itoa(pageSize))
	params.Set("po", "1")
	params.Set("np", "1")
	params.Set("fltt", "2")
	params.Set("invt", "2")
	params.Set("fid", "f3")
	params.Set("fs", "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23")
	params.Set("fields", "f12,f14,f2,f3")

	body, err := e.doGet(ctx, eastMoneyListPath, params)
	if err != nil {
		return nil, 0, err
	}

	var resp struct {
		Data struct {
			Total int `json:"total"`
			Diff  []struct {
				Code      string  `json:"f12"`
				Name      string  `json:"f14"`
				Price     float64 `json:"f2"`
				ChangePct float64 `json:"f3"`
			} `json:"diff"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}

	out := make([]StockBrief, 0, len(resp.Data.Diff))
	for _, row := range resp.Data.Diff {
		out = append(out, StockBrief{
			Code:      row.Code,
			Name:      row.Name,
			Price:     row.Price,
			ChangePct: row.ChangePct,
		})
	}
	return out, resp.Data.Total, nil
}

// Search 按代码或名称模糊搜索（拉取列表后本地过滤，开发阶段够用）
func (e *EastMoney) Search(ctx context.Context, keyword string, limit int) ([]StockBrief, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, nil
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	// 6 位纯数字直接查行情
	if len(keyword) == 6 {
		isDigit := true
		for _, c := range keyword {
			if c < '0' || c > '9' {
				isDigit = false
				break
			}
		}
		if isDigit {
			q, err := e.Quote(ctx, keyword)
			if err != nil {
				return nil, err
			}
			return []StockBrief{{
				Code:      q.Code,
				Name:      q.Name,
				Price:     q.Price,
				ChangePct: q.ChangePct,
			}}, nil
		}
	}

	items, _, err := e.List(ctx, 1, 100)
	if err != nil {
		return nil, err
	}
	kw := strings.ToLower(keyword)
	var out []StockBrief
	for _, s := range items {
		if strings.Contains(s.Code, keyword) ||
			strings.Contains(strings.ToLower(s.Name), kw) {
			out = append(out, s)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
