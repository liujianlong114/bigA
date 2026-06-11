package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/lijianjun/bigA/internal/config"
	"github.com/lijianjun/bigA/internal/orchestrator"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	root, _ := os.Getwd()
	if v := os.Getenv("PROJECT_ROOT"); v != "" {
		root = v
	}

	interval := 30 * time.Minute
	if v := os.Getenv("MARVIS_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	once := os.Getenv("MARVIS_ONCE") == "1" || os.Getenv("MARVIS_ONCE") == "true"

	orch := orchestrator.New(orchestrator.Config{
		ProjectRoot:  root,
		BaseURL:      getenv("MARVIS_BASE_URL", "http://localhost:8080"),
		FlutterURL:   getenv("MARVIS_FLUTTER_URL", "http://localhost:3000"),
		Interval:     interval,
		ReportDir:    filepath.Join(root, "marvis_promot"),
		BackendBin:   filepath.Join(root, "bin", "bigA"),
		StartFlutter: getenv("MARVIS_START_FLUTTER", "true") == "true",
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		cancel()
	}()

	log.Printf("marvis-orchestrator: MySQL=%s Redis=%s", mask(cfg.MySQLDSN), cfg.RedisAddr)

	if once {
		orch.RunOnce(ctx)
		return
	}
	orch.Start(ctx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mask(s string) string {
	if len(s) > 24 {
		return s[:24] + "..."
	}
	return s
}
