package sim

import (
	"context"

	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/store"
)

type PositionView struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Board       string  `json:"board"`
	Quantity    int     `json:"quantity"`
	Available   int     `json:"available"`
	CostPrice   float64 `json:"cost_price"`
	MarketPrice float64 `json:"market_price"`
	MarketValue float64 `json:"market_value"`
	CostValue   float64 `json:"cost_value"`
	Profit      float64 `json:"profit"`
	ProfitPct   float64 `json:"profit_pct"`
	ChangePct   float64 `json:"change_pct"`
	LimitUp     float64 `json:"limit_up"`
	LimitDown   float64 `json:"limit_down"`
	BuyDate     string  `json:"buy_date"`
}

type PortfolioSummary struct {
	AccountID      int64          `json:"account_id"`
	Cash           float64        `json:"cash"`
	FrozenCash     float64        `json:"frozen_cash"`
	MarketValue    float64        `json:"market_value"`
	TotalAssets    float64        `json:"total_assets"`
	TotalCost      float64        `json:"total_cost"`
	TotalProfit    float64        `json:"total_profit"`
	TotalProfitPct float64        `json:"total_profit_pct"`
	Positions      []PositionView `json:"positions"`
}

type Portfolio struct {
	market *market.Service
	repo   *store.Repo
}

func NewPortfolio(ms *market.Service, repo *store.Repo) *Portfolio {
	return &Portfolio{market: ms, repo: repo}
}

func (p *Portfolio) Summary(ctx context.Context, accountID int64) (*PortfolioSummary, error) {
	acct, err := p.repo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	positions, err := p.repo.ListPositions(ctx, accountID)
	if err != nil {
		return nil, err
	}

	s := &PortfolioSummary{
		AccountID:  accountID,
		Cash:       acct.Cash,
		FrozenCash: acct.FrozenCash,
		Positions:  make([]PositionView, 0, len(positions)),
	}

	for _, pos := range positions {
		q, err := p.market.GetQuote(ctx, pos.Code)
		if err != nil {
			continue
		}
		mktVal := round2(q.Price * float64(pos.Quantity))
		costVal := round2(pos.CostPrice * float64(pos.Quantity))
		profit := round2(mktVal - costVal)
		pct := 0.0
		if costVal > 0 {
			pct = round2(profit / costVal * 100)
		}
		s.Positions = append(s.Positions, PositionView{
			Code: pos.Code, Name: pos.Name, Board: pos.Board,
			Quantity: pos.Quantity, Available: pos.Available,
			CostPrice: pos.CostPrice, MarketPrice: q.Price,
			MarketValue: mktVal, CostValue: costVal,
			Profit: profit, ProfitPct: pct, ChangePct: q.ChangePct,
			LimitUp: q.LimitUp, LimitDown: q.LimitDown, BuyDate: pos.BuyDate,
		})
		s.MarketValue += mktVal
		s.TotalCost += costVal
	}

	s.MarketValue = round2(s.MarketValue)
	s.TotalCost = round2(s.TotalCost)
	s.TotalAssets = round2(s.Cash + s.MarketValue)
	s.TotalProfit = round2(s.MarketValue - s.TotalCost)
	if s.TotalCost > 0 {
		s.TotalProfitPct = round2(s.TotalProfit / s.TotalCost * 100)
	}
	return s, nil
}
