package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//go:embed migrations.sql
var migrationSQL embed.FS

// Open 连接 MySQL 并执行迁移
func Open(dsn string) (*sql.DB, error) {
	// 确保库存在
	if err := ensureDatabase(dsn); err != nil {
		return nil, err
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("mysql ping: %w", err)
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := applySchemaPatches(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func applySchemaPatches(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	patches := []string{
		`ALTER TABLE sim_account ADD COLUMN user_id BIGINT NULL`,
		`ALTER TABLE sim_account ADD UNIQUE KEY uk_user (user_id)`,
	}
	for _, p := range patches {
		if _, err := db.ExecContext(ctx, p); err != nil {
			msg := err.Error()
			if strings.Contains(msg, "Duplicate column") ||
				strings.Contains(msg, "Duplicate key name") ||
				strings.Contains(msg, "already exists") {
				continue
			}
			return fmt.Errorf("schema patch: %w", err)
		}
	}
	return nil
}

func ensureDatabase(dsn string) error {
	idx := strings.LastIndex(dsn, "/")
	if idx < 0 {
		return nil
	}
	base := dsn[:idx]
	rest := dsn[idx+1:]
	qIdx := strings.Index(rest, "?")
	dbName := rest
	if qIdx >= 0 {
		dbName = rest[:qIdx]
	}
	if dbName == "" {
		return nil
	}
	adminDSN := base + "/?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"
	if qIdx >= 0 {
		adminDSN = base + "/?" + rest[qIdx+1:]
	}
	db, err := sql.Open("mysql", adminDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS `"+dbName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	return err
}

func migrate(db *sql.DB) error {
	b, err := migrationSQL.ReadFile("migrations.sql")
	if err != nil {
		return err
	}
	// 去掉 CREATE DATABASE / USE，连接已指定库
	sqlText := string(b)
	sqlText = strings.ReplaceAll(sqlText, "CREATE DATABASE IF NOT EXISTS biga CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", "")
	sqlText = strings.ReplaceAll(sqlText, "USE biga;", "")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	for _, stmt := range splitSQL(sqlText) {
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w\nstmt: %s", err, truncate(stmt, 120))
		}
	}
	return nil
}

func splitSQL(s string) []string {
	var out []string
	var cur strings.Builder
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}
		cur.WriteString(line)
		cur.WriteString("\n")
		if strings.HasSuffix(line, ";") {
			out = append(out, strings.TrimSpace(cur.String()))
			cur.Reset()
		}
	}
	if t := strings.TrimSpace(cur.String()); t != "" {
		out = append(out, t)
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
