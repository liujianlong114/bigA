package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrUserExists = errors.New("username already exists")

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *Repo) GetUserByUsername(ctx context.Context, username string) (*User, string, error) {
	var u User
	var hash string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, nickname, password_hash, created_at FROM sim_user WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.Nickname, &hash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	return &u, hash, nil
}

func (r *Repo) GetUserByID(ctx context.Context, id int64) (*User, error) {
	var u User
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, nickname, created_at FROM sim_user WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.Nickname, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repo) GetAccountByUserID(ctx context.Context, userID int64) (*Account, error) {
	var a Account
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, cash, frozen_cash, updated_at FROM sim_account WHERE user_id = ?`, userID,
	).Scan(&a.ID, &a.Name, &a.Cash, &a.FrozenCash, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// RegisterUser 创建用户并绑定独立模拟账户（一用户一账户）
func (r *Repo) RegisterUser(ctx context.Context, username, passwordHash, nickname string, initialCash float64) (*User, *Account, error) {
	if nickname == "" {
		nickname = username
	}
	var user *User
	var acct *Account
	err := r.RunTx(ctx, func(tx *sql.Tx) error {
		var n int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM sim_user WHERE username = ?`, username).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			return ErrUserExists
		}
		now := time.Now()
		res, err := tx.ExecContext(ctx,
			`INSERT INTO sim_user (username, password_hash, nickname, created_at, updated_at) VALUES (?,?,?,?,?)`,
			username, passwordHash, nickname, now, now,
		)
		if err != nil {
			return err
		}
		uid, _ := res.LastInsertId()
		user = &User{ID: uid, Username: username, Nickname: nickname, CreatedAt: now}

		acctName := "u_" + username
		res, err = tx.ExecContext(ctx,
			`INSERT INTO sim_account (user_id, name, cash, frozen_cash, created_at, updated_at) VALUES (?,?,?,0,?,?)`,
			uid, acctName, initialCash, now, now,
		)
		if err != nil {
			return err
		}
		aid, _ := res.LastInsertId()
		acct = &Account{ID: aid, Name: acctName, Cash: initialCash, UpdatedAt: now}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return user, acct, nil
}
