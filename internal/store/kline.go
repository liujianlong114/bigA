package store

import (
	"context"
	"time"
)

// KlineBarRecord K 线落库记录
type KlineBarRecord struct {
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
	Amount float64
}

// UpsertKlines 批量写入 K 线（MySQL 落库）
func (r *Repo) UpsertKlines(ctx context.Context, code, period, source string, bars []KlineBarRecord) error {
	if len(bars) == 0 {
		return nil
	}
	now := time.Now()
	for _, b := range bars {
		td, err := time.Parse("2006-01-02", b.Date)
		if err != nil {
			continue
		}
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO market_kline (code, period, trade_date, open_p, high_p, low_p, close_p, volume, amount, source, recorded_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)
			ON DUPLICATE KEY UPDATE
				open_p=VALUES(open_p), high_p=VALUES(high_p), low_p=VALUES(low_p), close_p=VALUES(close_p),
				volume=VALUES(volume), amount=VALUES(amount), source=VALUES(source), recorded_at=VALUES(recorded_at)`,
			code, period, td.Format("2006-01-02"), b.Open, b.High, b.Low, b.Close, b.Volume, b.Amount, source, now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListKlines 从 MySQL 读 K 线
func (r *Repo) ListKlines(ctx context.Context, code, period string, limit int) ([]KlineBarRecord, error) {
	if limit <= 0 || limit > 500 {
		limit = 120
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT trade_date, open_p, high_p, low_p, close_p, volume, amount
		FROM market_kline WHERE code=? AND period=?
		ORDER BY trade_date DESC LIMIT ?`, code, period, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []KlineBarRecord
	for rows.Next() {
		var b KlineBarRecord
		var td time.Time
		if err := rows.Scan(&td, &b.Open, &b.High, &b.Low, &b.Close, &b.Volume, &b.Amount); err != nil {
			return nil, err
		}
		b.Date = td.Format("2006-01-02")
		out = append(out, b)
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, rows.Err()
}
