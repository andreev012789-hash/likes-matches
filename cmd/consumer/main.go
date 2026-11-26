package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/andreev012789-hash/likes-matches/internal/config"
	"github.com/andreev012789-hash/likes-matches/internal/event"
	"github.com/segmentio/kafka-go"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{cfg.KafkaBrokers},
		GroupID: "likes-consumer",
		Topic:   "events",
	})
	defer func() {
		if err := reader.Close(); err != nil {
			slog.Error("close kafka reader", "error", err)
		}
	}()

	slog.Info("consumer started", "group", "likes-consumer", "topic", "events")

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.Error("fetch message", "error", err)
			}
			break
		}

		var e event.Event
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			slog.Error("decode event", "partition", msg.Partition, "offset", msg.Offset, "error", err)
			continue
		}

		slog.Info("event received",
			"type", e.Type,
			"id", e.ID,
			"from_id", e.FromID,
			"to_id", e.ToID,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)
	}
	slog.Info("consumer stopped")
}
