package store

import (
	"context"
	"strings"
	"time"
)

// StockSnapshot 股票最新行情快照（全市场，非 tick 日志）
type StockSnapshot struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Board     string    `json:"board"`
	Price     float64   `json:"price"`
	ChangePct float64   `json:"change_pct"`
	PrevClose float64   `json:"prev_close"`
	Volume    int64     `json:"volume"`
	Amount    float64   `json:"amount"`
	LimitUp   float64   `json:"limit_up"`
	LimitDown float64   `json:"limit_down"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *Repo) UpsertStockSnapshot(ctx context.Context, s StockSnapshot) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO stock_snapshot (code, name, board, price, change_pct, prev_close, volume, amount, limit_up, limit_down, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
			name=VALUES(name), board=VALUES(board), price=VALUES(price), change_pct=VALUES(change_pct),
			prev_close=VALUES(prev_close), volume=VALUES(volume), amount=VALUES(amount),
			limit_up=VALUES(limit_up), limit_down=VALUES(limit_down), updated_at=VALUES(updated_at)`,
		s.Code, s.Name, s.Board, s.Price, s.ChangePct, s.PrevClose, s.Volume, s.Amount,
		s.LimitUp, s.LimitDown, time.Now(),
	)
	if err != nil {
		return err
	}
	_, _ = r.db.ExecContext(ctx,
		`INSERT INTO stock_info (code, name, board, updated_at) VALUES (?,?,?,?)
		 ON DUPLICATE KEY UPDATE name=VALUES(name), board=VALUES(board), updated_at=VALUES(updated_at)`,
		s.Code, s.Name, s.Board, time.Now(),
	)
	return nil
}

func (r *Repo) BatchUpsertStockSnapshots(ctx context.Context, rows []StockSnapshot) error {
	if len(rows) == 0 {
		return nil
	}
	const batch = 200
	now := time.Now()
	for i := 0; i < len(rows); i += batch {
		end := i + batch
		if end > len(rows) {
			end = len(rows)
		}
		if err := r.upsertSnapshotBatch(ctx, rows[i:end], now); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repo) upsertSnapshotBatch(ctx context.Context, rows []StockSnapshot, now time.Time) error {
	if len(rows) == 0 {
		return nil
	}
	var sb strings.Builder
	sb.WriteString(`INSERT INTO stock_snapshot (code,name,board,price,change_pct,prev_close,volume,amount,limit_up,limit_down,updated_at) VALUES `)
	args := make([]any, 0, len(rows)*11)
	for i, s := range rows {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString("(?,?,?,?,?,?,?,?,?,?,?)")
		args = append(args, s.Code, s.Name, s.Board, s.Price, s.ChangePct, s.PrevClose,
			s.Volume, s.Amount, s.LimitUp, s.LimitDown, now)
	}
	sb.WriteString(` ON DUPLICATE KEY UPDATE
		name=VALUES(name), board=VALUES(board), price=VALUES(price), change_pct=VALUES(change_pct),
		prev_close=VALUES(prev_close), volume=VALUES(volume), amount=VALUES(amount),
		limit_up=VALUES(limit_up), limit_down=VALUES(limit_down), updated_at=VALUES(updated_at)`)
	_, err := r.db.ExecContext(ctx, sb.String(), args...)
	return err
}

func (r *Repo) CountStockSnapshots(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM stock_snapshot`).Scan(&n)
	return n, err
}

func (r *Repo) ListStockSnapshots(ctx context.Context, board, keyword, sort string, page, pageSize int) ([]StockSnapshot, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	where := "WHERE 1=1"
	args := []any{}
	if board != "" && board != "all" {
		where += " AND board=?"
		args = append(args, board)
	}
	if kw := strings.TrimSpace(keyword); kw != "" {
		where += " AND (code LIKE ? OR name LIKE ?)"
		like := "%" + kw + "%"
		args = append(args, like, like)
	}
	order := "change_pct DESC"
	switch sort {
	case "change_pct_asc":
		order = "change_pct ASC"
	case "volume":
		order = "volume DESC"
	case "code":
		order = "code ASC"
	case "name":
		order = "name ASC"
	}
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM stock_snapshot "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	q := `SELECT code,name,board,price,change_pct,prev_close,volume,amount,limit_up,limit_down,updated_at
		FROM stock_snapshot ` + where + ` ORDER BY ` + order + ` LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []StockSnapshot
	for rows.Next() {
		var s StockSnapshot
		if err := rows.Scan(&s.Code, &s.Name, &s.Board, &s.Price, &s.ChangePct, &s.PrevClose,
			&s.Volume, &s.Amount, &s.LimitUp, &s.LimitDown, &s.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func (r *Repo) BoardStats(ctx context.Context) (map[string]int64, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT board, COUNT(*) FROM stock_snapshot GROUP BY board`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{"all": 0}
	for rows.Next() {
		var board string
		var n int64
		if err := rows.Scan(&board, &n); err != nil {
			return nil, err
		}
		out[board] = n
		out["all"] += n
	}
	return out, rows.Err()
}
