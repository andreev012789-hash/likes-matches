package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/andreev012789-hash/likes-matches/internal/config"
	"github.com/andreev012789-hash/likes-matches/internal/handler"
	"github.com/andreev012789-hash/likes-matches/internal/repository"
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

	logger.Info("connected to postage")

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func NewLogger(level string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
