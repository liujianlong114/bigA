package sim

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/store"
)

// Engine 大 A 模拟盘撮合引擎（MySQL 落库）
type Engine struct {
	market *market.Service
	repo   *store.Repo
	cache  *cache.Redis
	rules  *Rules
}

func NewEngine(ms *market.Service, repo *store.Repo, c *cache.Redis, rules *Rules) *Engine {
	return &Engine{market: ms, repo: repo, cache: c, rules: rules}
}

type OrderRequest struct {
	Code      string   `json:"code"`
	Side      string   `json:"side"`       // buy | sell
	OrderType string   `json:"order_type"` // market | limit
	Price     *float64 `json:"price,omitempty"`
	Quantity  int      `json:"quantity"`
	Source    string   `json:"source,omitempty"`
	Remark    string   `json:"remark,omitempty"`
}

type OrderResult struct {
	OrderID      int64   `json:"order_id"`
	TradeID      int64   `json:"trade_id,omitempty"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Side         string  `json:"side"`
	OrderType    string  `json:"order_type"`
	Status       string  `json:"status"`
	FillPrice    float64 `json:"fill_price,omitempty"`
	Quantity     int     `json:"quantity"`
	Amount       float64 `json:"amount,omitempty"`
	Commission   float64 `json:"commission,omitempty"`
	StampTax     float64 `json:"stamp_tax,omitempty"`
	TransferFee  float64 `json:"transfer_fee,omitempty"`
	CashAfter    float64 `json:"cash_after,omitempty"`
	RejectReason string  `json:"reject_reason,omitempty"`
}

func (e *Engine) PlaceOrder(ctx context.Context, accountID int64, req OrderRequest) (*OrderResult, error) {
	unlock, err := e.cache.LockTrade(ctx, accountID, 15*time.Second)
	if err != nil {
		return nil, err
	}
	defer unlock()

	code := normalizeCode(req.Code)
	if code == "" {
		return nil, fmt.Errorf("invalid code")
	}
	side := strings.ToLower(req.Side)
	orderType := strings.ToLower(req.OrderType)
	if side != "buy" && side != "sell" {
		return nil, fmt.Errorf("side must be buy or sell")
	}
	if orderType == "" {
		orderType = "market"
	}
	if orderType != "market" && orderType != "limit" {
		return nil, fmt.Errorf("order_type must be market or limit")
	}
	source := req.Source
	if source == "" {
		source = "ai"
	}

	q, err := e.market.GetQuoteForce(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("quote: %w", err)
	}

	if err := e.rules.ValidateOrder(side, orderType, req.Quantity, req.Price, q); err != nil {
		return e.rejectOrder(ctx, accountID, req, code, q.Name, side, orderType, source, err.Error())
	}

	fillPrice, err := MatchPrice(side, orderType, req.Price, q)
	if err != nil {
		return e.rejectOrder(ctx, accountID, req, code, q.Name, side, orderType, source, err.Error())
	}

	amount := round2(fillPrice * float64(req.Quantity))
	commission := CalcCommission(amount)
	stampTax := 0.0
	transferFee := CalcTransferFee(amount)
	if side == "sell" {
		stampTax = CalcStampTax(amount)
	}

	tradeDate := time.Now().Format("2006-01-02")
	today := tradeDate
	board := string(market.DetectBoard(code, q.Name))

	var result OrderResult

	err = e.repo.RunTx(ctx, func(tx *sql.Tx) error {
		cash, err := e.repo.GetAccountCashTx(ctx, tx, accountID)
		if err != nil {
			return err
		}

		pos, err := e.repo.GetPositionTx(ctx, tx, accountID, code)
		if err != nil {
			return err
		}

		switch side {
		case "buy":
			totalCost := amount + commission + transferFee
			if cash < totalCost {
				return fmt.Errorf("可用资金不足: 需要 %.2f, 可用 %.2f", totalCost, cash)
			}
			cash -= totalCost

			var qty, avail int
			var cost float64
			var buyDate string
			if pos != nil {
				qty = pos.Quantity + req.Quantity
				cost = round2((pos.CostPrice*float64(pos.Quantity) + amount) / float64(qty))
				avail = pos.Available
				buyDate = pos.BuyDate
			} else {
				qty = req.Quantity
				cost = fillPrice
				avail = 0
				buyDate = today
			}
			if err := e.repo.UpsertPositionTx(ctx, tx, &store.Position{
				AccountID: accountID, Code: code, Name: q.Name, Board: board,
				Quantity: qty, Available: avail, CostPrice: cost, BuyDate: buyDate,
			}); err != nil {
				return err
			}

		case "sell":
			if pos == nil || pos.Available < req.Quantity {
				avail := 0
				if pos != nil {
					avail = pos.Available
				}
				return fmt.Errorf("可卖数量不足(T+1): 需要 %d, 可卖 %d", req.Quantity, avail)
			}
			proceeds := amount - commission - stampTax - transferFee
			cash += proceeds

			qty := pos.Quantity - req.Quantity
			avail := pos.Available - req.Quantity
			if qty == 0 {
				if err := e.repo.DeletePositionTx(ctx, tx, accountID, code); err != nil {
					return err
				}
			} else {
				if err := e.repo.UpsertPositionTx(ctx, tx, &store.Position{
					AccountID: accountID, Code: code, Name: q.Name, Board: board,
					Quantity: qty, Available: avail, CostPrice: pos.CostPrice, BuyDate: pos.BuyDate,
				}); err != nil {
					return err
				}
			}
		}

		if err := e.repo.UpdateAccountCashTx(ctx, tx, accountID, cash); err != nil {
			return err
		}

		order := &store.Order{
			AccountID: accountID, Code: code, Name: q.Name, Side: side, OrderType: orderType,
			Price: req.Price, Quantity: req.Quantity, FilledQty: req.Quantity,
			Status: "filled", Source: source, Remark: req.Remark,
		}
		orderID, err := insertOrderTx(ctx, tx, order)
		if err != nil {
			return err
		}

		trade := &store.Trade{
			AccountID: accountID, OrderID: orderID, Code: code, Name: q.Name, Side: side,
			Price: fillPrice, Quantity: req.Quantity, Amount: amount,
			Commission: commission, StampTax: stampTax, TransferFee: transferFee,
			TradeDate: tradeDate, TradedAt: time.Now(),
		}
		tradeID, err := insertTradeTx(ctx, tx, trade)
		if err != nil {
			return err
		}

		result = OrderResult{
			OrderID: orderID, TradeID: tradeID, Code: code, Name: q.Name,
			Side: side, OrderType: orderType, Status: "filled", FillPrice: fillPrice,
			Quantity: req.Quantity, Amount: amount, Commission: commission,
			StampTax: stampTax, TransferFee: transferFee, CashAfter: cash,
		}
		return nil
	})
	if err != nil {
		_ = e.repo.LogAIAction(ctx, accountID, "place_order", req, nil, false, err.Error())
		return nil, err
	}
	_ = e.repo.LogAIAction(ctx, accountID, "place_order", req, result, true, "")
	return &result, nil
}

func (e *Engine) rejectOrder(ctx context.Context, accountID int64, req OrderRequest, code, name, side, orderType, source, reason string) (*OrderResult, error) {
	o := &store.Order{
		AccountID: accountID, Code: code, Name: name, Side: side, OrderType: orderType,
		Price: req.Price, Quantity: req.Quantity, Status: "rejected", RejectReason: reason, Source: source,
	}
	id, err := e.repo.InsertOrder(ctx, o)
	if err != nil {
		return nil, err
	}
	_ = e.repo.LogAIAction(ctx, accountID, "place_order", req, map[string]any{"order_id": id, "reason": reason}, false, reason)
	return &OrderResult{OrderID: id, Code: code, Name: name, Side: side, OrderType: orderType, Status: "rejected", RejectReason: reason}, fmt.Errorf("%s", reason)
}

func (e *Engine) SettleDay(ctx context.Context, accountID int64) (int64, error) {
	td := time.Now().Format("2006-01-02")
	n, err := e.repo.SettleT1(ctx, accountID, td)
	if err != nil {
		return 0, err
	}
	_ = e.repo.LogAIAction(ctx, accountID, "settle_day", nil, map[string]int64{"positions_settled": n}, true, "")
	return n, nil
}

func insertOrderTx(ctx context.Context, tx *sql.Tx, o *store.Order) (int64, error) {
	now := time.Now()
	var price interface{}
	if o.Price != nil {
		price = *o.Price
	}
	res, err := tx.ExecContext(ctx, `
		INSERT INTO sim_order (account_id, code, name, side, order_type, price, quantity, filled_qty, status, reject_reason, source, remark, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		o.AccountID, o.Code, o.Name, o.Side, o.OrderType, price, o.Quantity, o.FilledQty, o.Status,
		nullIfEmpty(o.RejectReason), o.Source, nullIfEmpty(o.Remark), now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func insertTradeTx(ctx context.Context, tx *sql.Tx, t *store.Trade) (int64, error) {
	res, err := tx.ExecContext(ctx, `
		INSERT INTO sim_trade (account_id, order_id, code, name, side, price, quantity, amount, commission, stamp_tax, transfer_fee, trade_date, traded_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		t.AccountID, t.OrderID, t.Code, t.Name, t.Side, t.Price, t.Quantity, t.Amount,
		t.Commission, t.StampTax, t.TransferFee, t.TradeDate, t.TradedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func normalizeCode(code string) string {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return ""
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return ""
		}
	}
	return code
}
