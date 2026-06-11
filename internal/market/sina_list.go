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
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// SinaList 新浪财经 — A 股列表/搜索（push2 不可用时的主力源）
type SinaList struct {
	client *http.Client
}

func NewSinaList() *SinaList {
	return &SinaList{client: &http.Client{Timeout: 15 * time.Second}}
}

func (s *SinaList) fetchJSON(ctx context.Context, rawURL string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, dest); err == nil {
		return nil
	}
	utf8, _, err := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), body)
	if err != nil {
		return json.Unmarshal(body, dest)
	}
	return json.Unmarshal(utf8, dest)
}

// List 沪深 A 股分页
func (s *SinaList) List(ctx context.Context, page, pageSize int) ([]StockBrief, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	raw := fmt.Sprintf(
		"https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?page=%d&num=%d&sort=changepercent&asc=0&node=hs_a",
		page, pageSize,
	)
	var rows []struct {
		Code          string    `json:"code"`
		Name          string    `json:"name"`
		Trade         string    `json:"trade"`
		ChangePercent flexFloat `json:"changepercent"`
	}
	if err := s.fetchJSON(ctx, raw, &rows); err != nil {
		return nil, 0, err
	}
	out := make([]StockBrief, 0, len(rows))
	for _, r := range rows {
		price, _ := strconv.ParseFloat(r.Trade, 64)
		pct := float64(r.ChangePercent)
		out = append(out, StockBrief{
			Code:      r.Code,
			Name:      r.Name,
			Price:     price,
			ChangePct: pct,
		})
	}
	// 新浪不返 total，A 股约 5500
	return out, 5500, nil
}

// Search 代码/名称/拼音
func (s *SinaList) Search(ctx context.Context, keyword string, limit int) ([]StockBrief, error) {
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		return nil, nil
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}
	raw := "https://suggest3.sinajs.cn/suggest/type=11,12,13,14,15&key=" + url.QueryEscape(keyword)
	var payload string
	if err := s.fetchSuggest(ctx, raw, &payload); err != nil {
		return nil, err
	}
	if payload == "" {
		return nil, nil
	}
	var out []StockBrief
	for _, seg := range strings.Split(payload, ";") {
		parts := strings.Split(seg, ",")
		if len(parts) < 4 {
			continue
		}
		code := parts[2]
		if len(code) != 6 {
			continue
		}
		name := parts[3]
		q, err := NewSina().Quote(ctx, code)
		if err != nil {
			out = append(out, StockBrief{Code: code, Name: name})
		} else {
			out = append(out, StockBrief{Code: code, Name: q.Name, Price: q.Price, ChangePct: q.ChangePct})
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *SinaList) fetchSuggest(ctx context.Context, rawURL string, out *string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	line := string(body)
	utf8, _, _ := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), body)
	if strings.Contains(string(utf8), "suggestvalue") {
		line = string(utf8)
	}
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")
	if start < 0 || end <= start {
		return nil
	}
	*out = line[start+1 : end]
	return nil
}
