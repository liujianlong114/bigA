package market

import (
	"context"
	"encoding/json"

	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/store"
)

// Service 行情：Redis 缓存 + MySQL 全量落库
type Service struct {
	provider *Provider
	cache    *cache.Redis
	repo     *store.Repo
}

func NewService(p *Provider, c *cache.Redis, r *store.Repo) *Service {
	return &Service{provider: p, cache: c, repo: r}
}

// GetQuote 拉取真实行情，写 Redis + MySQL
func (s *Service) GetQuote(ctx context.Context, code string) (*Quote, error) {
	var cached Quote
	if s.cache.GetJSON(ctx, code, &cached) && cached.Price > 0 {
		return &cached, nil
	}
	q, err := s.provider.Quote(ctx, code)
	if err != nil {
		return nil, err
	}
	_ = s.cache.SetJSON(ctx, code, q)
	s.persistQuote(ctx, q)
	return q, nil
}

// GetQuoteForce 强制刷新并落库（AI 下单前建议调用）
func (s *Service) GetQuoteForce(ctx context.Context, code string) (*Quote, error) {
	q, err := s.provider.Quote(ctx, code)
	if err != nil {
		return nil, err
	}
	_ = s.cache.SetJSON(ctx, code, q)
	s.persistQuote(ctx, q)
	return q, nil
}

func (s *Service) persistQuote(ctx context.Context, q *Quote) {
	payload, _ := json.Marshal(q)
	_, _ = s.repo.InsertQuoteLog(ctx, store.QuoteRecord{
		Code: q.Code, Name: q.Name, Source: q.Source,
		Price: q.Price, Open: q.Open, High: q.High, Low: q.Low,
		PrevClose: q.PrevClose, ChangePct: q.ChangePct, Volume: q.Volume, Amount: q.Amount,
		Bid1: q.Bid1, Ask1: q.Ask1, LimitUp: q.LimitUp, LimitDown: q.LimitDown,
		PayloadJSON: payload,
	})
	_ = s.repo.UpsertStockInfo(ctx, q.Code, q.Name, q.Board)
}

func (s *Service) List(ctx context.Context, page, pageSize int) ([]StockBrief, int, error) {
	return s.provider.List(ctx, page, pageSize)
}

func (s *Service) Search(ctx context.Context, keyword string, limit int) ([]StockBrief, error) {
	return s.provider.Search(ctx, keyword, limit)
}

func (s *Service) QuoteHistory(ctx context.Context, code string, limit int) ([]store.QuoteLog, error) {
	return s.repo.ListQuoteLogs(ctx, code, limit)
}

// GetKline 拉取 K 线并落库 MySQL（支持全 A 股任意代码）
func (s *Service) GetKline(ctx context.Context, code string, period KlinePeriod, limit int) (*KlineResult, error) {
	if period == "" {
		period = PeriodDay
	}
	res, err := s.provider.Kline(ctx, code, period, limit)
	if err != nil {
		return nil, err
	}
	bars := make([]store.KlineBarRecord, len(res.Bars))
	for i, b := range res.Bars {
		bars[i] = store.KlineBarRecord{
			Date: b.Date, Open: b.Open, High: b.High, Low: b.Low, Close: b.Close, Volume: b.Volume, Amount: b.Amount,
		}
	}
	_ = s.repo.UpsertKlines(ctx, res.Code, string(period), res.Source, bars)
	if res.Name != "" {
		_ = s.repo.UpsertStockInfo(ctx, res.Code, res.Name, string(DetectBoard(res.Code, res.Name)))
	}
	return res, nil
}

func (s *Service) KlineFromDB(ctx context.Context, code string, period KlinePeriod, limit int) ([]KlineBar, error) {
	rows, err := s.repo.ListKlines(ctx, code, string(period), limit)
	if err != nil {
		return nil, err
	}
	out := make([]KlineBar, len(rows))
	for i, b := range rows {
		out[i] = KlineBar{Date: b.Date, Open: b.Open, High: b.High, Low: b.Low, Close: b.Close, Volume: b.Volume, Amount: b.Amount}
	}
	return out, nil
}
