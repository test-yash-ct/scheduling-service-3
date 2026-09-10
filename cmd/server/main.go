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
	"github.com/healthops/scheduling-service/internal/config"
	"github.com/healthops/scheduling-service/internal/events"
	"github.com/healthops/scheduling-service/internal/handlers"
	"github.com/healthops/scheduling-service/internal/obs"
	"github.com/healthops/scheduling-service/internal/service"
	"github.com/healthops/scheduling-service/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer pool.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(obs.RequestID())
	r.Use(obs.AccessLogger())

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	r.GET("/meta", obs.MetaHandler(cfg.Metadata))

	v1 := r.Group("/v1")
	st := store.New(pool)
	outbox := events.NewMemory()
	booking := service.NewBooking(st, outbox)
	(&handlers.AppointmentAPI{Store: st, Booking: booking}).Register(v1)
	(&handlers.AdminAPI{Store: st}).Register(v1)

	srv := &http.Server{Addr: cfg.ListenAddr, Handler: r}
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
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
