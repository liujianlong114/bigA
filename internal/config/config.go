package config

import (
	"os"
	"strconv"
)

type Config struct {
	MySQLDSN      string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	HTTPAddr      string
	InitialCash   float64
	RelaxHours    bool
	JWTSecret     string
	AuthRelax     bool
}

func Load() Config {
	return Config{
		MySQLDSN:      getenv("MYSQL_DSN", "root@tcp(127.0.0.1:3306)/biga?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"),
		RedisAddr:     getenv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		RedisDB:       getenvInt("REDIS_DB", 0),
		HTTPAddr:      getenv("HTTP_ADDR", ":8080"),
		InitialCash:   getenvFloat("SIM_INITIAL_CASH", 1_000_000),
		RelaxHours:    getenvBool("SIM_RELAX_HOURS", true),
		JWTSecret:     getenv("JWT_SECRET", "biga-dev-secret-change-me"),
		AuthRelax:     getenvBool("SIM_AUTH_RELAX", true),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func getenvFloat(k string, def float64) float64 {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
}

func getenvBool(k string, def bool) bool {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v == "1" || v == "true" || v == "TRUE" || v == "yes"
}
