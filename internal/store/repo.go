package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Repo MySQL 持久化（模拟盘全量落库）
type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

type Account struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Cash       float64   `json:"cash"`
	FrozenCash float64   `json:"frozen_cash"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Position struct {
	ID        int64     `json:"id"`
	AccountID int64     `json:"account_id"`
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Board     string    `json:"board"`
	Quantity  int       `json:"quantity"`
	Available int       `json:"available"`
	FrozenQty int       `json:"frozen_qty"`
	CostPrice float64   `json:"cost_price"`
	BuyDate   string    `json:"buy_date"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Order struct {
	ID           int64     `json:"id"`
	AccountID    int64     `json:"account_id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Side         string    `json:"side"`
	OrderType    string    `json:"order_type"`
	Price        *float64  `json:"price,omitempty"`
	Quantity     int       `json:"quantity"`
	FilledQty    int       `json:"filled_qty"`
	Status       string    `json:"status"`
	RejectReason string    `json:"reject_reason,omitempty"`
	Source       string    `json:"source"`
	Remark       string    `json:"remark,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Trade struct {
	ID          int64     `json:"id"`
	AccountID   int64     `json:"account_id"`
	OrderID     int64     `json:"order_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Side        string    `json:"side"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	Amount      float64   `json:"amount"`
	Commission  float64   `json:"commission"`
	StampTax    float64   `json:"stamp_tax"`
	TransferFee float64   `json:"transfer_fee"`
	TradeDate   string    `json:"trade_date"`
	TradedAt    time.Time `json:"traded_at"`
}

type QuoteLog struct {
	ID         int64     `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Source     string    `json:"source"`
	Price      float64   `json:"price"`
	ChangePct  float64   `json:"change_pct"`
	LimitUp    float64   `json:"limit_up"`
	LimitDown  float64   `json:"limit_down"`
	RecordedAt time.Time `json:"recorded_at"`
}

func (r *Repo) EnsureDefaultAccount(ctx context.Context, initialCash float64) (*Account, error) {
	var a Account
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, cash, frozen_cash, updated_at FROM sim_account WHERE name = 'default'`,
	).Scan(&a.ID, &a.Name, &a.Cash, &a.FrozenCash, &a.UpdatedAt)
	if err == nil {
		return &a, nil
	}
	if err != sql.ErrNoRows {
		return nil, err
	}
	now := time.Now()
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO sim_account (name, cash, frozen_cash, created_at, updated_at) VALUES ('default', ?, 0, ?, ?)`,
		initialCash, now, now,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return &Account{ID: id, Name: "default", Cash: initialCash, UpdatedAt: now}, nil
}

func (r *Repo) GetAccount(ctx context.Context, id int64) (*Account, error) {
	var a Account
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, cash, frozen_cash, updated_at FROM sim_account WHERE id = ?`, id,
	).Scan(&a.ID, &a.Name, &a.Cash, &a.FrozenCash, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repo) ListPositions(ctx context.Context, accountID int64) ([]Position, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, account_id, code, name, board, quantity, available, frozen_qty, cost_price, buy_date, updated_at
		 FROM sim_position WHERE account_id = ? ORDER BY code`, accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Position
	for rows.Next() {
		var p Position
		var buyDate time.Time
		if err := rows.Scan(&p.ID, &p.AccountID, &p.Code, &p.Name, &p.Board, &p.Quantity, &p.Available,
			&p.FrozenQty, &p.CostPrice, &buyDate, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.BuyDate = buyDate.Format("2006-01-02")
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repo) GetPosition(ctx context.Context, accountID int64, code string) (*Position, error) {
	var p Position
	var buyDate time.Time
	err := r.db.QueryRowContext(ctx,
		`SELECT id, account_id, code, name, board, quantity, available, frozen_qty, cost_price, buy_date, updated_at
		 FROM sim_position WHERE account_id = ? AND code = ?`, accountID, code,
	).Scan(&p.ID, &p.AccountID, &p.Code, &p.Name, &p.Board, &p.Quantity, &p.Available,
		&p.FrozenQty, &p.CostPrice, &buyDate, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.BuyDate = buyDate.Format("2006-01-02")
	return &p, nil
}

// QuoteRecord 行情落库记录
type QuoteRecord struct {
	Code, Name, Source                           string
	Price, Open, High, Low, PrevClose, ChangePct float64
	Volume                                       int64
	Amount, Bid1, Ask1, LimitUp, LimitDown       float64
	PayloadJSON                                  []byte
}

func (r *Repo) InsertQuoteLog(ctx context.Context, q QuoteRecord) (int64, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO market_quote_log
		(code, name, source, price, open_p, high_p, low_p, prev_close, change_pct, volume, amount,
		 bid1, ask1, limit_up, limit_down, payload_json, recorded_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		q.Code, q.Name, q.Source, q.Price, q.Open, q.High, q.Low, q.PrevClose, q.ChangePct,
		q.Volume, q.Amount, q.Bid1, q.Ask1, q.LimitUp, q.LimitDown, q.PayloadJSON, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) ListQuoteLogs(ctx context.Context, code string, limit int) ([]QuoteLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, code, name, source, price, change_pct, limit_up, limit_down, recorded_at
		FROM market_quote_log WHERE code = ? ORDER BY id DESC LIMIT ?`, code, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []QuoteLog
	for rows.Next() {
		var l QuoteLog
		if err := rows.Scan(&l.ID, &l.Code, &l.Name, &l.Source, &l.Price, &l.ChangePct,
			&l.LimitUp, &l.LimitDown, &l.RecordedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *Repo) UpsertStockInfo(ctx context.Context, code, name, board string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO stock_info (code, name, board, updated_at) VALUES (?,?,?,?)
		ON DUPLICATE KEY UPDATE name=VALUES(name), board=VALUES(board), updated_at=VALUES(updated_at)`,
		code, name, board, time.Now(),
	)
	return err
}

func (r *Repo) InsertOrder(ctx context.Context, o *Order) (int64, error) {
	now := time.Now()
	var price interface{}
	if o.Price != nil {
		price = *o.Price
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO sim_order (account_id, code, name, side, order_type, price, quantity, filled_qty, status, reject_reason, source, remark, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		o.AccountID, o.Code, o.Name, o.Side, o.OrderType, price, o.Quantity, o.FilledQty, o.Status, nullStr(o.RejectReason), o.Source, nullStr(o.Remark), now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) UpdateOrder(ctx context.Context, o *Order) error {
	var price interface{}
	if o.Price != nil {
		price = *o.Price
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE sim_order SET filled_qty=?, status=?, reject_reason=?, price=?, updated_at=?
		WHERE id=?`, o.FilledQty, o.Status, nullStr(o.RejectReason), price, time.Now(), o.ID,
	)
	return err
}

func (r *Repo) InsertTrade(ctx context.Context, t *Trade) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
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

func (r *Repo) UpdateAccountCash(ctx context.Context, accountID int64, cash float64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE sim_account SET cash=?, updated_at=? WHERE id=?`, cash, time.Now(), accountID,
	)
	return err
}

func (r *Repo) UpsertPosition(ctx context.Context, p *Position) error {
	buyDate, _ := time.Parse("2006-01-02", p.BuyDate)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sim_position (account_id, code, name, board, quantity, available, frozen_qty, cost_price, buy_date, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
			name=VALUES(name), board=VALUES(board), quantity=VALUES(quantity), available=VALUES(available),
			frozen_qty=VALUES(frozen_qty), cost_price=VALUES(cost_price), buy_date=VALUES(buy_date), updated_at=VALUES(updated_at)`,
		p.AccountID, p.Code, p.Name, p.Board, p.Quantity, p.Available, p.FrozenQty, p.CostPrice, buyDate, time.Now(),
	)
	return err
}

func (r *Repo) DeletePosition(ctx context.Context, accountID int64, code string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM sim_position WHERE account_id=? AND code=?`, accountID, code,
	)
	return err
}

func (r *Repo) SettleT1(ctx context.Context, accountID int64, tradeDate string) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE sim_position SET available = quantity, updated_at = ? WHERE account_id = ?`, time.Now(), accountID,
	)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	detail, _ := json.Marshal(map[string]any{"rows": n})
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO sim_settlement_log (account_id, trade_date, action, detail_json, created_at) VALUES (?,?,?,?,?)`,
		accountID, tradeDate, "t1_settle", detail, time.Now(),
	)
	return n, err
}

func (r *Repo) ListOrders(ctx context.Context, accountID int64, limit int) ([]Order, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, account_id, code, name, side, order_type, price, quantity, filled_qty, status,
			COALESCE(reject_reason,''), source, COALESCE(remark,''), created_at, updated_at
		FROM sim_order WHERE account_id = ? ORDER BY id DESC LIMIT ?`, accountID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *Repo) ListTrades(ctx context.Context, accountID int64, limit int) ([]Trade, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, account_id, order_id, code, name, side, price, quantity, amount, commission, stamp_tax, transfer_fee, trade_date, traded_at
		FROM sim_trade WHERE account_id = ? ORDER BY id DESC LIMIT ?`, accountID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trade
	for rows.Next() {
		var t Trade
		var td time.Time
		if err := rows.Scan(&t.ID, &t.AccountID, &t.OrderID, &t.Code, &t.Name, &t.Side, &t.Price, &t.Quantity,
			&t.Amount, &t.Commission, &t.StampTax, &t.TransferFee, &td, &t.TradedAt); err != nil {
			return nil, err
		}
		t.TradeDate = td.Format("2006-01-02")
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *Repo) LogAIAction(ctx context.Context, accountID int64, action string, req, resp any, success bool, errMsg string) error {
	reqJ, _ := json.Marshal(req)
	respJ, _ := json.Marshal(resp)
	ok := 0
	if success {
		ok = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ai_action_log (account_id, action, request_json, response_json, success, error_msg, created_at)
		VALUES (?,?,?,?,?,?,?)`,
		accountID, action, reqJ, respJ, ok, nullStr(errMsg), time.Now(),
	)
	return err
}

func (r *Repo) RunTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repo) GetPositionTx(ctx context.Context, tx *sql.Tx, accountID int64, code string) (*Position, error) {
	var p Position
	var buyDate time.Time
	err := tx.QueryRowContext(ctx,
		`SELECT id, account_id, code, name, board, quantity, available, frozen_qty, cost_price, buy_date, updated_at
		 FROM sim_position WHERE account_id = ? AND code = ?`, accountID, code,
	).Scan(&p.ID, &p.AccountID, &p.Code, &p.Name, &p.Board, &p.Quantity, &p.Available,
		&p.FrozenQty, &p.CostPrice, &buyDate, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.BuyDate = buyDate.Format("2006-01-02")
	return &p, nil
}

func (r *Repo) UpdateAccountCashTx(ctx context.Context, tx *sql.Tx, accountID int64, cash float64) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE sim_account SET cash=?, updated_at=? WHERE id=?`, cash, time.Now(), accountID,
	)
	return err
}

func (r *Repo) UpsertPositionTx(ctx context.Context, tx *sql.Tx, p *Position) error {
	buyDate, _ := time.Parse("2006-01-02", p.BuyDate)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO sim_position (account_id, code, name, board, quantity, available, frozen_qty, cost_price, buy_date, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
			name=VALUES(name), board=VALUES(board), quantity=VALUES(quantity), available=VALUES(available),
			frozen_qty=VALUES(frozen_qty), cost_price=VALUES(cost_price), updated_at=VALUES(updated_at)`,
		p.AccountID, p.Code, p.Name, p.Board, p.Quantity, p.Available, p.FrozenQty, p.CostPrice, buyDate, time.Now(),
	)
	return err
}

func (r *Repo) DeletePositionTx(ctx context.Context, tx *sql.Tx, accountID int64, code string) error {
	_, err := tx.ExecContext(ctx,
		`DELETE FROM sim_position WHERE account_id=? AND code=?`, accountID, code,
	)
	return err
}

func (r *Repo) GetAccountCashTx(ctx context.Context, tx *sql.Tx, accountID int64) (float64, error) {
	var cash float64
	err := tx.QueryRowContext(ctx, `SELECT cash FROM sim_account WHERE id=?`, accountID).Scan(&cash)
	return cash, err
}

func scanOrders(rows *sql.Rows) ([]Order, error) {
	var out []Order
	for rows.Next() {
		var o Order
		var price sql.NullFloat64
		if err := rows.Scan(&o.ID, &o.AccountID, &o.Code, &o.Name, &o.Side, &o.OrderType, &price,
			&o.Quantity, &o.FilledQty, &o.Status, &o.RejectReason, &o.Source, &o.Remark, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		if price.Valid {
			v := price.Float64
			o.Price = &v
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// AIState 供 AI 一次性读取的完整状态
type AIState struct {
	Account   *Account   `json:"account"`
	Positions []Position `json:"positions"`
	Orders    []Order    `json:"recent_orders"`
	Trades    []Trade    `json:"recent_trades"`
	Timestamp time.Time  `json:"timestamp"`
}

func (r *Repo) BuildAIState(ctx context.Context, accountID int64) (*AIState, error) {
	acct, err := r.GetAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	pos, err := r.ListPositions(ctx, accountID)
	if err != nil {
		return nil, err
	}
	orders, err := r.ListOrders(ctx, accountID, 20)
	if err != nil {
		return nil, err
	}
	trades, err := r.ListTrades(ctx, accountID, 20)
	if err != nil {
		return nil, err
	}
	return &AIState{
		Account: acct, Positions: pos, Orders: orders, Trades: trades, Timestamp: time.Now(),
	}, nil
}

func (r *Repo) DB() *sql.DB { return r.db }

func (r *Repo) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repo) Stats(ctx context.Context) (map[string]int64, error) {
	tables := []string{"sim_account", "sim_position", "sim_order", "sim_trade", "market_quote_log", "market_kline", "ai_action_log"}
	out := make(map[string]int64)
	for _, t := range tables {
		var n int64
		if err := r.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", t)).Scan(&n); err != nil {
			return nil, err
		}
		out[t] = n
	}
	return out, nil
}
