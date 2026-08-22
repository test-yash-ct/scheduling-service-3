package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/healthops/scheduling-service/internal/audit"
	"github.com/healthops/scheduling-service/internal/config"
	"github.com/healthops/scheduling-service/internal/handlers"
	"github.com/healthops/scheduling-service/internal/middleware"
	"github.com/healthops/scheduling-service/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.RequestLogger())

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	v1 := r.Group("/v1")
	v1.Use(middleware.Authenticate(cfg.JWTSecret, cfg.MaxTokenTTLSec))
	v1.Use(middleware.RateLimit(60, time.Minute))
	st := store.New(pool)
	al := audit.New()
	(&handlers.AppointmentAPI{Store: st, Audit: al}).Register(v1)
	admin := v1.Group("")
	admin.Use(middleware.RequireCSRF(cfg.CSRFCookieName, cfg.JWTSecret))
	(&handlers.AdminAPI{Store: st, Audit: al}).Register(admin)
	(&handlers.NotifyAPI{Allowlist: cfg.WebhookAllowlist}).Register(v1)

	srv := &http.Server{Addr: cfg.ListenAddr, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("listening on %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}
