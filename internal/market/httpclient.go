package market

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// resilientGet 依次尝试多个 URL（应对 push2 在某些网络被 reset）
func resilientGet(ctx context.Context, client *http.Client, urls []string, headers map[string]string) ([]byte, error) {
	var lastErr error
	for _, raw := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
		if err != nil {
			lastErr = err
			continue
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			continue
		}
		if len(body) == 0 {
			lastErr = fmt.Errorf("empty body")
			continue
		}
		return body, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("all upstream urls failed")
	}
	return nil, lastErr
}

func defaultHeaders() map[string]string {
	return map[string]string{
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Referer":    "https://quote.eastmoney.com/",
	}
}

func buildURL(base string, params url.Values) string {
	u, _ := url.Parse(base)
	u.RawQuery = params.Encode()
	return u.String()
}

func eastmoneyHosts(pathWithQuery func(host string) string) []string {
	hosts := []string{
		"https://push2his.eastmoney.com",
		"http://push2his.eastmoney.com",
		"https://push2.eastmoney.com",
		"http://push2.eastmoney.com",
		"http://82.push2.eastmoney.com",
		"https://15.push2.eastmoney.com",
	}
	out := make([]string, 0, len(hosts))
	for _, h := range hosts {
		out = append(out, pathWithQuery(h))
	}
	return out
}

func trimBOM(s string) string {
	return strings.TrimPrefix(s, "\ufeff")
}
