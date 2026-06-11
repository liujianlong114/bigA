package store

import (
	"context"
	"database/sql"
	"time"
)

// ConditionalOrder 条件单
type ConditionalOrder struct {
	ID             int64      `json:"id"`
	AccountID      int64      `json:"account_id"`
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	ConditionType  string     `json:"condition_type"`
	TriggerValue   float64    `json:"trigger_value"`
	Side           string     `json:"side"`
	OrderType      string     `json:"order_type"`
	ActionPrice    *float64   `json:"price,omitempty"`
	Quantity       int        `json:"quantity"`
	Status         string     `json:"status"`
	TriggerOrderID *int64     `json:"trigger_order_id,omitempty"`
	TriggerMessage string     `json:"trigger_message,omitempty"`
	ValidDate      string     `json:"valid_date"`
	Source         string     `json:"source"`
	Remark         string     `json:"remark,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	TriggeredAt    *time.Time `json:"triggered_at,omitempty"`
}

func (r *Repo) InsertConditionalOrder(ctx context.Context, co *ConditionalOrder) (int64, error) {
	now := time.Now()
	var actionPrice interface{}
	if co.ActionPrice != nil {
		actionPrice = *co.ActionPrice
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO sim_conditional_order
		(account_id, code, name, condition_type, trigger_value, side, order_type, action_price,
		 quantity, status, valid_date, source, remark, created_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		co.AccountID, co.Code, co.Name, co.ConditionType, co.TriggerValue, co.Side, co.OrderType,
		actionPrice, co.Quantity, co.Status, co.ValidDate, co.Source, nullStr(co.Remark), now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repo) ListConditionalOrders(ctx context.Context, accountID int64, limit int) ([]ConditionalOrder, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, account_id, code, name, condition_type, trigger_value, side, order_type,
			action_price, quantity, status, trigger_order_id, COALESCE(trigger_message,''),
			valid_date, source, COALESCE(remark,''), created_at, updated_at, triggered_at
		FROM sim_conditional_order WHERE account_id = ? ORDER BY id DESC LIMIT ?`, accountID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConditionalOrders(rows)
}

func (r *Repo) ListPendingConditionalOrders(ctx context.Context, validDate string) ([]ConditionalOrder, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, account_id, code, name, condition_type, trigger_value, side, order_type,
			action_price, quantity, status, trigger_order_id, COALESCE(trigger_message,''),
			valid_date, source, COALESCE(remark,''), created_at, updated_at, triggered_at
		FROM sim_conditional_order WHERE status = 'pending' AND valid_date = ?`, validDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanConditionalOrders(rows)
}

func (r *Repo) CancelConditionalOrder(ctx context.Context, accountID, id int64) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE sim_conditional_order SET status='cancelled', updated_at=? 
		WHERE id=? AND account_id=? AND status='pending'`, time.Now(), id, accountID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repo) MarkConditionalTriggered(ctx context.Context, id int64) (bool, error) {
	now := time.Now()
	res, err := r.db.ExecContext(ctx, `
		UPDATE sim_conditional_order SET status='triggering', triggered_at=?, updated_at=?
		WHERE id=? AND status='pending'`, now, now, id,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *Repo) UpdateConditionalOrderResult(ctx context.Context, id int64, status string, orderID *int64, msg string) error {
	var oid interface{}
	if orderID != nil {
		oid = *orderID
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE sim_conditional_order SET status=?, trigger_order_id=?, trigger_message=?, updated_at=?
		WHERE id=?`, status, oid, nullStr(msg), time.Now(), id,
	)
	return err
}

func (r *Repo) ExpireConditionalOrders(ctx context.Context, today string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sim_conditional_order SET status='expired', updated_at=?
		WHERE status='pending' AND valid_date < ?`, time.Now(), today,
	)
	return err
}

func scanConditionalOrders(rows *sql.Rows) ([]ConditionalOrder, error) {
	var out []ConditionalOrder
	for rows.Next() {
		var co ConditionalOrder
		var actionPrice sql.NullFloat64
		var triggerOrderID sql.NullInt64
		var triggeredAt sql.NullTime
		var validDate time.Time
		if err := rows.Scan(
			&co.ID, &co.AccountID, &co.Code, &co.Name, &co.ConditionType, &co.TriggerValue,
			&co.Side, &co.OrderType, &actionPrice, &co.Quantity, &co.Status, &triggerOrderID,
			&co.TriggerMessage, &validDate, &co.Source, &co.Remark, &co.CreatedAt, &co.UpdatedAt, &triggeredAt,
		); err != nil {
			return nil, err
		}
		if actionPrice.Valid {
			v := actionPrice.Float64
			co.ActionPrice = &v
		}
		if triggerOrderID.Valid {
			v := triggerOrderID.Int64
			co.TriggerOrderID = &v
		}
		if triggeredAt.Valid {
			t := triggeredAt.Time
			co.TriggeredAt = &t
		}
		co.ValidDate = validDate.Format("2006-01-02")
		out = append(out, co)
	}
	return out, rows.Err()
}
