package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lijianjun/bigA/internal/analysis"
	"github.com/lijianjun/bigA/internal/api"
	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/config"
	"github.com/lijianjun/bigA/internal/db"
	"github.com/lijianjun/bigA/internal/market"
	"github.com/lijianjun/bigA/internal/sim"
	"github.com/lijianjun/bigA/internal/store"
	"github.com/lijianjun/bigA/internal/stream"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	sectorSvc := market.NewSectorService(market.NewEastMoney(), rdb)
	rules := sim.NewRules(cfg.RelaxHours)
	engine := sim.NewEngine(mktSvc, repo, rdb, rules)
	condMgr := sim.NewConditionalManager(engine, repo, mktSvc)
	portfolio := sim.NewPortfolio(mktSvc, repo)
	perfSvc := sim.NewPerformance(repo, mktSvc, portfolio, cfg.InitialCash)

	universe := market.NewUniverse()
	hub := stream.NewHub()
	streamEngine := stream.NewEngine(universe, rdb, hub, cfg.RelaxHours)
	streamEngine.SetConditionalChecker(condMgr)
	streamEngine.Start(ctx)

	go func() {
		log.Printf("stream: 加载全 A 股列表...")
		err := universe.LoadAll(ctx, rdb,
			market.NewSinaList(),
			market.NewEastMoney(),
		)
		if err != nil {
			log.Printf("stream: universe 加载失败（将重试）: %v", err)
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if universe.Size() > 0 {
						return
					}
					if err := universe.LoadAll(ctx, rdb, market.NewSinaList(), market.NewEastMoney()); err == nil {
						log.Printf("stream: universe 重试成功 %d 只", universe.Size())
						return
					}
				}
			}
		}
	}()

	wsHandler := &api.WSHandler{Hub: hub, Redis: rdb}

	ana := analysis.New(sqlDB, rdb.Client())

	h := &api.Handlers{
		AccountID:      acct.ID,
		Market:         &api.MarketAdapter{Svc: mktSvc},
		SectorSvc:      sectorSvc,
		Engine:         engine,
		Conditional:    condMgr,
		PerformanceSvc: perfSvc,
		PortfolioSvc:   portfolio,
		Repo:           repo,
		LiveRedis:      rdb,
		Analyzer:       ana,
	}

	log.Printf("bigA 大A模拟盘启动 %s", cfg.HTTPAddr)
	log.Printf("MySQL: %s | Redis: %s | 账户 id=%d 现金=%.0f", maskDSN(cfg.MySQLDSN), cfg.RedisAddr, acct.ID, acct.Cash)
	log.Printf("实时行情: WS /api/v1/ws/market | REST /api/v1/market/live")
	log.Printf("AI: GET /api/v1/ai/state  POST /api/v1/ai/order")

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	if err := http.ListenAndServe(cfg.HTTPAddr, api.NewRouter(h, wsHandler)); err != nil {
		log.Fatal(err)
	}
}

func maskDSN(dsn string) string {
	if i := len(dsn); i > 20 {
		return dsn[:20] + "..."
	}
	return dsn
}
