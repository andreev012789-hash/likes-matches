package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andreev012789-hash/likes-matches/internal/config"
	"github.com/andreev012789-hash/likes-matches/internal/handler"
	"github.com/andreev012789-hash/likes-matches/internal/producer"
	"github.com/andreev012789-hash/likes-matches/internal/ratelimit"
	"github.com/andreev012789-hash/likes-matches/internal/repository"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	logger := NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	logger.Info("starting", "http_addr", cfg.HTTPAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("connected to postgres")

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Error("connect to redis", "error", err)
		os.Exit(1)
	}

	defer func() {
		if err := rdb.Close(); err != nil {
			slog.Error("redis close failed", "error", err)
		}
	}()

	limiter := ratelimit.New(rdb, 30, time.Minute)

	prod := producer.New([]string{cfg.KafkaBrokers}, "events")
	defer func() {
		if err := prod.Close(); err != nil {
			slog.Error("kafka close failed", "error", err)
		}
	}()

	profileRepo := repository.NewProfileRepositore(pool)
	likeRepo := repository.NewLikeRepository(pool)
	h := handler.New(profileRepo, likeRepo, limiter)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           h.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		logger.Error("server failed", "error", err)
		os.Exit(1)
	case <-sigCtx.Done():
		logger.Info("shutting down")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", "error", err)
	}

	logger.Info("server stopped")
}

func NewLogger(level string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
