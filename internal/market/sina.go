package market

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Sina 新浪财经免费行情（备用）
type Sina struct {
	client *http.Client
}

func NewSina() *Sina {
	return &Sina{client: &http.Client{Timeout: 12 * time.Second}}
}

func (s *Sina) Quote(ctx context.Context, code string) (*Quote, error) {
	sym := SinaSymbol(code)
	if sym == "" {
		return nil, fmt.Errorf("invalid stock code: %s", code)
	}
	url := "https://hq.sinajs.cn/list=" + sym
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://finance.sina.com.cn")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return nil, err
	}

	utf8, _, err := transform.Bytes(simplifiedchinese.GBK.NewDecoder(), body)
	if err != nil {
		utf8 = body
	}
	line := string(utf8)
	// var hq_str_sh600519="名称,今开,昨收,现价,...";
	start := strings.Index(line, "\"")
	end := strings.LastIndex(line, "\"")
	if start < 0 || end <= start {
		return nil, fmt.Errorf("sina: parse failed for %s", code)
	}
	fields := strings.Split(line[start+1:end], ",")
	if len(fields) < 10 {
		return nil, fmt.Errorf("sina: insufficient fields for %s", code)
	}

	parse := func(i int) float64 {
		v, _ := strconv.ParseFloat(fields[i], 64)
		return v
	}

	name := fields[0]
	open := parse(1)
	prev := parse(2)
	price := parse(3)
	high := parse(4)
	low := parse(5)
	vol, _ := strconv.ParseInt(fields[8], 10, 64)
	amount, _ := strconv.ParseFloat(fields[9], 64)

	change := price - prev
	pct := 0.0
	if prev > 0 {
		pct = change / prev * 100
	}

	bid1, ask1 := 0.0, 0.0
	if len(fields) > 21 {
		bid1 = parse(11)
		ask1 = parse(21)
	}

	return &Quote{
		Code:      code,
		Name:      name,
		Price:     price,
		Open:      open,
		High:      high,
		Low:       low,
		PrevClose: prev,
		Change:    change,
		ChangePct: pct,
		Volume:    vol / 100,
		Amount:    amount,
		Bid1:      bid1,
		Ask1:      ask1,
		Source:    "sina",
		UpdatedAt: time.Now(),
	}, nil
}
