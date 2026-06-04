package market

import "context"

// Provider 行情数据源（东方财富优先，新浪兜底）
type Provider struct {
	primary  *EastMoney
	fallback *Sina
}

func NewProvider() *Provider {
	return &Provider{
		primary:  NewEastMoney(),
		fallback: NewSina(),
	}
}

func (p *Provider) Quote(ctx context.Context, code string) (*Quote, error) {
	q, err := p.primary.Quote(ctx, code)
	if err == nil && q.Price > 0 {
		EnrichQuote(q)
		return q, nil
	}
	q, err = p.fallback.Quote(ctx, code)
	if err != nil {
		return nil, err
	}
	EnrichQuote(q)
	return q, nil
}

func (p *Provider) List(ctx context.Context, page, pageSize int) ([]StockBrief, int, error) {
	return p.primary.List(ctx, page, pageSize)
}

func (p *Provider) Search(ctx context.Context, keyword string, limit int) ([]StockBrief, error) {
	return p.primary.Search(ctx, keyword, limit)
}
