package dbpool

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}
	cfg.MaxConns = int32(getenvInt("DB_MAX_CONNS", 25))
	cfg.MinConns = int32(getenvInt("DB_MIN_CONNS", 2))
	cfg.MaxConnLifetime = time.Duration(getenvInt("DB_MAX_CONN_LIFETIME_SEC", 1800)) * time.Second
	cfg.MaxConnIdleTime = time.Duration(getenvInt("DB_MAX_CONN_IDLE_SEC", 300)) * time.Second
	cfg.HealthCheckPeriod = 30 * time.Second
	return pgxpool.NewWithConfig(ctx, cfg)
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
