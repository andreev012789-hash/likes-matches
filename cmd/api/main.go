package main

import (
	"log/slog"
	"os"

	"github.com/andreev012789-hash/likes-matches/internal/config"
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
}

func NewLogger(level string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(handler)
}
