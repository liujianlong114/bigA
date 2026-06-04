package main

import (
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/lijianjun/bigA/internal/api"
	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/config"
	"github.com/lijianjun/bigA/internal/db"
	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/sim"
	"github.com/lijianjun/bigA/internal/store"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	ctx := context.Background()

	sqlDB, err := db.Open(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	defer sqlDB.Close()

	rdb, err := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer rdb.Close()

	repo := store.NewRepo(sqlDB)
	acct, err := repo.EnsureDefaultAccount(ctx, cfg.InitialCash)
	if err != nil {
		log.Fatalf("account: %v", err)
	}

	provider := market.NewProvider()
	mktSvc := market.NewService(provider, rdb, repo)
	rules := sim.NewRules(cfg.RelaxHours)
	engine := sim.NewEngine(mktSvc, repo, rdb, rules)
	portfolio := sim.NewPortfolio(mktSvc, repo)

	h := &api.Handlers{
		AccountID:    acct.ID,
		Market:       &api.MarketAdapter{Svc: mktSvc},
		Engine:       engine,
		PortfolioSvc: portfolio,
		Repo:         repo,
	}

	log.Printf("bigA 大A模拟盘启动 %s", cfg.HTTPAddr)
	log.Printf("MySQL: %s | Redis: %s | 账户 id=%d 现金=%.0f", maskDSN(cfg.MySQLDSN), cfg.RedisAddr, acct.ID, acct.Cash)
	log.Printf("AI: GET /api/v1/ai/state  POST /api/v1/ai/order")

	if err := http.ListenAndServe(cfg.HTTPAddr, api.NewRouter(h)); err != nil {
		log.Fatal(err)
	}
}

func maskDSN(dsn string) string {
	if i := len(dsn); i > 20 {
		return dsn[:20] + "..."
	}
	return dsn
}
