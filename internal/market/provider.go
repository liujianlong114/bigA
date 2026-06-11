package market

import "context"

// Provider 行情数据源（多源容错）
type Provider struct {
	primary  *EastMoney
	fallback *Sina
	sinaList *SinaList
	kline    *KlineFetcher
}

func NewProvider() *Provider {
	return &Provider{
		primary:  NewEastMoney(),
		fallback: NewSina(),
		sinaList: NewSinaList(),
		kline:    NewKlineFetcher(),
	}
}

func (p *Provider) Quote(ctx context.Context, code string) (*Quote, error) {
	// 优先新浪：push2 在很多网络不可用
	q, err := p.fallback.Quote(ctx, code)
	if err == nil && q.Price > 0 {
		EnrichQuote(q)
		return q, nil
	}
	q, err = p.primary.Quote(ctx, code)
	if err != nil {
		return nil, err
	}
	EnrichQuote(q)
	return q, nil
}

func (p *Provider) List(ctx context.Context, page, pageSize int) ([]StockBrief, int, error) {
	items, total, err := p.sinaList.List(ctx, page, pageSize)
	if err == nil && len(items) > 0 {
		return items, total, nil
	}
	return p.primary.List(ctx, page, pageSize)
}

func (p *Provider) Search(ctx context.Context, keyword string, limit int) ([]StockBrief, error) {
	items, err := p.sinaList.Search(ctx, keyword, limit)
	if err == nil && len(items) > 0 {
		return items, nil
	}
	return p.primary.Search(ctx, keyword, limit)
}

func (p *Provider) Kline(ctx context.Context, code string, period KlinePeriod, limit int) (*KlineResult, error) {
	return p.kline.Fetch(ctx, code, period, limit)
}
