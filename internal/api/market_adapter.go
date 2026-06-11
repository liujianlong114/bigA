package api

import (
	"net/http"
	"strconv"

	"github.com/lijianjun/bigA/internal/market"
)

type MarketAdapter struct {
	Svc *market.Service
}

func (a *MarketAdapter) GetQuote(r *http.Request, code string) (any, error) {
	return a.Svc.GetQuote(r.Context(), code)
}

func (a *MarketAdapter) List(r *http.Request) (any, error) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	items, total, err := a.Svc.List(r.Context(), page, size)
	if err != nil {
		return nil, err
	}
	return map[string]any{"total": total, "items": items}, nil
}

func (a *MarketAdapter) Search(r *http.Request, q string) (any, error) {
	items, err := a.Svc.Search(r.Context(), q, 20)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items}, nil
}

func (a *MarketAdapter) QuoteHistory(r *http.Request, code string, limit int) (any, error) {
	return a.Svc.QuoteHistory(r.Context(), code, limit)
}

func (a *MarketAdapter) Kline(r *http.Request, code, period string, limit int) (any, error) {
	return a.Svc.GetKline(r.Context(), code, market.KlinePeriod(period), limit)
}
