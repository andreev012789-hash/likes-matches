package config

import (
	"errors"
	"os"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	LogLevel     string
	RedisAddr    string
	KafkaBrokers string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:     os.Getenv("HTTP_ADDR"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
		RedisAddr:    os.Getenv("REDIS_ADDR"),
		KafkaBrokers: os.Getenv("KAFKA_BROKERS"),
	}

	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = ":8080"
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
	if cfg.RedisAddr == "" {
		cfg.RedisAddr = "localhost:6380"
	}
	if cfg.KafkaBrokers == "" {
		cfg.KafkaBrokers = "localhost:9094"
	}
	return cfg, nil
}
