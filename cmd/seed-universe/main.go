package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lijianjun/bigA/internal/cache"
	"github.com/lijianjun/bigA/internal/config"
	"github.com/lijianjun/bigA/internal/market"
)

// 一次性把全 A 股代码表写入 Redis，避免启动时反复拉列表触发新浪限流
func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	ctx := context.Background()

	rdb, err := cache.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatal(err)
	}
	defer rdb.Close()

	u := market.NewUniverse()
	log.Println("拉取全 A 股列表（约 1-2 分钟）...")
	if err := u.LoadAll(ctx, rdb, market.NewSinaList(), market.NewEastMoney()); err != nil {
		log.Fatal(err)
	}
	log.Printf("完成：%d 只股票已缓存到 Redis key biga:universe:v1", u.Size())
	os.Exit(0)
}
