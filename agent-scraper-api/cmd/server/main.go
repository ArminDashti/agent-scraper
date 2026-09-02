package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArminDashti/agent-scraper-api/internal/auth"
	"github.com/ArminDashti/agent-scraper-api/internal/config"
	"github.com/ArminDashti/agent-scraper-api/internal/forwarder"
	httpserver "github.com/ArminDashti/agent-scraper-api/internal/http"
	"github.com/ArminDashti/agent-scraper-api/internal/pipeline"
	"github.com/ArminDashti/agent-scraper-api/internal/scheduler"
	"github.com/ArminDashti/agent-scraper-api/internal/scraper"
	"github.com/ArminDashti/agent-scraper-api/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()
	db, err := store.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres required: %v", err)
	}
	defer db.Close()

	authSvc := auth.NewService(db, cfg.JWTSecret)
	if err := authSvc.EnsureDefaultUser(ctx, cfg.DefaultUsername, cfg.DefaultPassword); err != nil {
		log.Fatalf("seed default user: %v", err)
	}

	extractor := scraper.NewExtractor()
	fwd := forwarder.NewClient(db, cfg.TargetAPIURL)
	runner := pipeline.NewRunner(db, extractor, fwd)

	sched, err := scheduler.Start(cfg.CronSchedule, runner)
	if err != nil {
		log.Fatalf("start schedule: %v", err)
	}
	defer sched.Stop()

	srv := httpserver.New(cfg, authSvc, db, runner, cfg.CronSchedule)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("agent-scraper-api listening on http://%s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}
