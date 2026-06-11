package sim

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/store"
)

// 条件类型
const (
	CondPriceGTE     = "price_gte"
	CondPriceLTE     = "price_lte"
	CondChangePctGTE = "change_pct_gte"
	CondChangePctLTE = "change_pct_lte"
)

// ConditionalOrderRequest 创建条件单
type ConditionalOrderRequest struct {
	Code          string   `json:"code"`
	ConditionType string   `json:"condition_type"`
	TriggerValue  float64  `json:"trigger_value"`
	Side          string   `json:"side"`
	OrderType     string   `json:"order_type"`
	Price         *float64 `json:"price,omitempty"`
	Quantity      int      `json:"quantity"`
	Remark        string   `json:"remark,omitempty"`
	Source        string   `json:"source,omitempty"`
}

// ConditionalManager 条件单：创建 / 取消 / 每秒 tick 检查触发
type ConditionalManager struct {
	engine *Engine
	repo   *store.Repo
	market *market.Service
}

func NewConditionalManager(eng *Engine, repo *store.Repo, ms *market.Service) *ConditionalManager {
	return &ConditionalManager{engine: eng, repo: repo, market: ms}
}

func (m *ConditionalManager) Create(ctx context.Context, accountID int64, req ConditionalOrderRequest) (*store.ConditionalOrder, error) {
	code := normalizeCode(req.Code)
	if code == "" {
		return nil, fmt.Errorf("invalid code")
	}
	ct := strings.ToLower(strings.TrimSpace(req.ConditionType))
	if !isValidConditionType(ct) {
		return nil, fmt.Errorf("condition_type must be price_gte|price_lte|change_pct_gte|change_pct_lte")
	}
	side := strings.ToLower(req.Side)
	if side != "buy" && side != "sell" {
		return nil, fmt.Errorf("side must be buy or sell")
	}
	orderType := strings.ToLower(req.OrderType)
	if orderType == "" {
		orderType = "market"
	}
	if orderType != "market" && orderType != "limit" {
		return nil, fmt.Errorf("order_type must be market or limit")
	}
	if orderType == "limit" && (req.Price == nil || *req.Price <= 0) {
		return nil, fmt.Errorf("limit order requires price")
	}
	if req.Quantity <= 0 || req.Quantity%LotSize != 0 {
		return nil, fmt.Errorf("quantity must be positive multiple of %d", LotSize)
	}

	q, err := m.market.GetQuote(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("quote: %w", err)
	}

	source := req.Source
	if source == "" {
		source = "api"
	}

	co := &store.ConditionalOrder{
		AccountID:     accountID,
		Code:          code,
		Name:          q.Name,
		ConditionType: ct,
		TriggerValue:  req.TriggerValue,
		Side:          side,
		OrderType:     orderType,
		ActionPrice:   req.Price,
		Quantity:      req.Quantity,
		Status:        "pending",
		ValidDate:     time.Now().Format("2006-01-02"),
		Source:        source,
		Remark:        req.Remark,
	}
	id, err := m.repo.InsertConditionalOrder(ctx, co)
	if err != nil {
		return nil, err
	}
	co.ID = id
	co.CreatedAt = time.Now()
	co.UpdatedAt = co.CreatedAt
	return co, nil
}

func (m *ConditionalManager) List(ctx context.Context, accountID int64, limit int) ([]store.ConditionalOrder, error) {
	return m.repo.ListConditionalOrders(ctx, accountID, limit)
}

func (m *ConditionalManager) Cancel(ctx context.Context, accountID, id int64) error {
	return m.repo.CancelConditionalOrder(ctx, accountID, id)
}

// CheckTriggers 由 stream 每秒 tick 调用
func (m *ConditionalManager) CheckTriggers(ctx context.Context, quotes []market.LiveQuote) {
	if len(quotes) == 0 {
		return
	}
	today := time.Now().Format("2006-01-02")
	_ = m.repo.ExpireConditionalOrders(ctx, today)

	quoteMap := make(map[string]market.LiveQuote, len(quotes))
	for _, q := range quotes {
		quoteMap[q.Code] = q
	}

	pending, err := m.repo.ListPendingConditionalOrders(ctx, today)
	if err != nil {
		log.Printf("conditional: list pending: %v", err)
		return
	}
	for _, co := range pending {
		q, ok := quoteMap[co.Code]
		if !ok || q.Price <= 0 {
			continue
		}
		if !conditionMet(co.ConditionType, co.TriggerValue, q.Price, q.ChangePct) {
			continue
		}
		m.trigger(ctx, co)
	}
}

func (m *ConditionalManager) trigger(ctx context.Context, co store.ConditionalOrder) {
	ok, err := m.repo.MarkConditionalTriggered(ctx, co.ID)
	if err != nil || !ok {
		return
	}

	res, err := m.engine.PlaceOrder(ctx, co.AccountID, OrderRequest{
		Code:      co.Code,
		Side:      co.Side,
		OrderType: co.OrderType,
		Price:     co.ActionPrice,
		Quantity:  co.Quantity,
		Source:    "conditional",
		Remark:    fmt.Sprintf("triggered from conditional #%d", co.ID),
	})
	if err != nil {
		reason := err.Error()
		if res != nil && res.RejectReason != "" {
			reason = res.RejectReason
		}
		_ = m.repo.UpdateConditionalOrderResult(ctx, co.ID, "failed", nil, reason)
		log.Printf("conditional: #%d trigger failed: %s", co.ID, reason)
		return
	}
	var orderID *int64
	if res != nil {
		orderID = &res.OrderID
	}
	_ = m.repo.UpdateConditionalOrderResult(ctx, co.ID, "triggered", orderID, "")
	log.Printf("conditional: #%d triggered code=%s side=%s order=%v", co.ID, co.Code, co.Side, orderID)
}

func isValidConditionType(ct string) bool {
	switch ct {
	case CondPriceGTE, CondPriceLTE, CondChangePctGTE, CondChangePctLTE:
		return true
	}
	return false
}

func conditionMet(condType string, trigger, price, changePct float64) bool {
	switch condType {
	case CondPriceGTE:
		return price >= trigger
	case CondPriceLTE:
		return price <= trigger
	case CondChangePctGTE:
		return changePct >= trigger
	case CondChangePctLTE:
		return changePct <= trigger
	}
	return false
}
